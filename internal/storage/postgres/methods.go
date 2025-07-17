package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/transactions_api/internal/models"
)

// интерфейс используется для передачи методам для работы с БД либо подключение к БД либо транзакцию
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

var (
	ErrLowBalance         = errors.New("low balance")
	ErrWalletDoesNotExist = errors.New("wallet with this address does not exist")
)

func MakeTransaction(ctx context.Context, trans *models.Transaction) error {

	// проверим существование кошельков с адресами отправителя и получателя
	if err := checkWalletExists(ctx, trans.SenderAddress); err != nil {
		return fmt.Errorf("[GetBalanceOfWallet]: %w", err)
	}
	if err := checkWalletExists(ctx, trans.RecieverAddress); err != nil {
		return fmt.Errorf("[GetBalanceOfWallet]: %w", err)
	}

	// проверим баланс отправителя
	if err := checkBalanceOfSenderWallet(ctx, trans); err != nil {
		return fmt.Errorf("[MakeTransaction|check balance]: %w", err)
	}

	// начало транзакции
	tx, err := Db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("[MakeTransaction|start transaction]: %w", err)
	}

	// откат при ошибке транзакции
	defer func() error {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				return fmt.Errorf("[MakeTransaction|rollback]: %w", rbErr)
			}
		}
		return nil
	}()

	// у отправителя отнимем деньги
	_, err = tx.Exec(ctx, "UPDATE wallets SET balance = balance - $1 WHERE address = $2", trans.Amount, trans.SenderAddress)
	if err != nil {
		return fmt.Errorf("[MakeTransaction|exec update balace from sender request]: %w", err)
	}
	// у получателя прибавим деньги
	_, err = tx.Exec(ctx, "UPDATE wallets SET balance = balance + $1 WHERE address = $2", trans.Amount, trans.RecieverAddress)
	if err != nil {
		return fmt.Errorf("[MakeTransaction|exec update balace from reciever request]: %w", err)
	}

	// добавим транзакцию в таблицу
	if err := addTransactionInTable(ctx, tx, trans); err != nil {
		return fmt.Errorf("[MakeTransaction|AddTransactionInTable]: %w", err)
	}

	// фиксация транзакции
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[MakeTransaction|commit transaction]: %w", err)
	}
	return nil
}

func addTransactionInTable(ctx context.Context, db DB, trans *models.Transaction) error {
	// получим id кошельков отправителя и получателя
	idSender, err := getWalletIdByAddress(ctx, db, trans.SenderAddress)
	if err != nil {
		return fmt.Errorf("[AddTransactionInTable]: %w", err)
	}
	idReceiver, err := getWalletIdByAddress(ctx, db, trans.RecieverAddress)
	if err != nil {
		return fmt.Errorf("[AddTransactionInTable]: %w", err)
	}

	// добавим транзакцию в таблицу транзакций
	_, err = db.Exec(ctx, "INSERT INTO transactions (sender_id,receiver_id,amount) VALUES($1,$2,$3)", idSender, idReceiver, trans.Amount)
	if err != nil {
		return fmt.Errorf("[AddTransactionInTable|exec set new transaction request]: %w", err)
	}

	return nil
}

func GetListOfLastTransactions(ctx context.Context, count int, trans *[]*models.Transaction) error {
	rows, err := Db.Query(ctx, `SELECT w1.address, w2.address , transactions.amount , transactions.date
								FROM transactions
								JOIN wallets AS w1 ON sender_id = w1.wallet_id
								JOIN wallets AS w2 ON receiver_id = w2.wallet_id 
								ORDER BY date DESC
								LIMIT $1`, count)
	if err != nil {
		return fmt.Errorf("[GetListOfLastTransactions|exec get all transaction]: %w", err)
	}

	for rows.Next() {
		var curTrans models.Transaction
		if err := rows.Scan(&curTrans.SenderAddress, &curTrans.RecieverAddress, &curTrans.Amount, &curTrans.Date); err != nil {
			return fmt.Errorf("[GetListOfLastTransactions|exec get transaction]: %w", err)
		}
		*trans = append(*trans, &curTrans)
	}

	return nil
}

func GetBalanceOfWallet(ctx context.Context, addr string) (float64, error) {

	// проверим существование кошелька с таким адресом
	if err := checkWalletExists(ctx, addr); err != nil {
		return -1, fmt.Errorf("[GetBalanceOfWallet]: %w", err)
	}

	var balance float64
	err := Db.QueryRow(ctx, "SELECT balance FROM wallets WHERE address = $1", addr).Scan(&balance)
	if err != nil {
		return -1, fmt.Errorf("[GetBalanceOfWallet|exec get balance request]: %w", err)
	}

	return balance, nil
}

func checkBalanceOfSenderWallet(ctx context.Context, trans *models.Transaction) error {
	balance, err := GetBalanceOfWallet(ctx, trans.SenderAddress)
	if err != nil {
		return fmt.Errorf("[CheckBalanceOfSenderWallet|exec get balance]: %w", err)
	}

	if balance < trans.Amount {
		return fmt.Errorf("[CheckBalanceOfSenderWallet]: %w", ErrLowBalance)
	}

	return nil
}

func checkWalletExists(ctx context.Context, addr string) error {
	var exists bool
	if err := Db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM wallets WHERE address = $1)", addr).Scan(&exists); err != nil {
		return fmt.Errorf("[CheckWalletExists|exec check exists]: %w", err)
	}

	if !exists {
		return fmt.Errorf("[CheckWalletExists]: %w", ErrWalletDoesNotExist)
	}

	return nil
}

func getWalletIdByAddress(ctx context.Context, db DB, addr string) (int, error) {
	var id int
	err := db.QueryRow(ctx, "SELECT wallet_id FROM wallets WHERE address = $1", addr).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("[GetWalletIdByAddress|exec get wallet_id]: %w", err)
	}

	return id, nil
}
