package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/transactions_api/internal/config"
)

var Db *pgx.Conn

func ConnectDB(cfg *config.Config) *pgx.Conn {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.User, cfg.Storage.Password, cfg.Storage.Name)

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("[ConnectDB|connection] ", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("[ConnectDB|ping]", err)
	}

	Db = conn
	return Db
}

func FillDatabase() {
	if err := createTables(Db); err != nil {
		log.Fatal(err)
	}

	if err := fillTableWallets(Db, 10); err != nil {
		log.Fatal(err)
	}
}

func createTables(db *pgx.Conn) error {
	data, err := os.ReadFile("./internal/storage/postgres/create_schema.sql")
	if err != nil {
		return fmt.Errorf("[createTables|read .sql file]: %w", err)
	}

	_, err = db.Exec(context.Background(), string(data))
	if err != nil {
		return fmt.Errorf("[createTables|exec create request]: %w", err)
	}

	return nil
}

func fillTableWallets(db *pgx.Conn, count int) error {
	// проверим, cозданы ли уже кошельки
	var exists bool
	if err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM wallets)").Scan(&exists); err != nil {
		return fmt.Errorf("[fillTableWallets|exec check exists]: %w", err)
	}
	if exists {
		return nil
	}
	// сгенерируем кошельки
	for i := 1; i <= count; i++ {

		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			return fmt.Errorf("[fillTableWallets|generate address]: %w", err)
		}
		addr := hex.EncodeToString(bytes)
		if _, err := Db.Exec(context.Background(), "INSERT INTO wallets (address) VALUES ($1)", addr); err != nil {
			return fmt.Errorf("[fillTableWallets|insert generated address]: %w", err)
		}
	}
	return nil
}
