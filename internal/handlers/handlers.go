package handlers

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/transactions_api/internal/models"
	"github.com/transactions_api/internal/storage/postgres"
)

// Send godoc
// @Summary Совершить перевод
// @Description Совершает транзакцию перевода средств с одного кошелька на другой
// @Tags Transactions
// @Accept json
// @Produce json
// @Param task body models.Transaction true "Данные транзакции"
// @Success 200 {object} map[string]interface{} "Транзакция успешно совершена"
// @Failure 400 {object} map[string]interface{} "'error': 'message'"
// @Failure 500 {object}  map[string]interface{} "'error': 'message'"
// @Router /api/send [post]
func Send(c *fiber.Ctx) error {
	var transaction models.Transaction

	// парсим JSON в структуру transaction
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный формат данных"})
	}

	// проверим корректность суммы перевода
	if transaction.Amount < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Сумма перевода не может быть отрицательной"})
	}

	// исполнение транзакции
	err := postgres.MakeTransaction(context.Background(), &transaction)
	if err != nil {
		if errors.Is(err, postgres.ErrWalletDoesNotExist) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный адрес кошелька"})
		}

		if errors.Is(err, postgres.ErrLowBalance) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Недостаточно средств для перевода"})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// успешное выполнение транзакции
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Транзакция успешно совершена"})
}

// GetLast godoc
// @Summary Получить последние совершенные транзакции
// @Description Возвращает последние n совершенных транзакций
// @Tags Transactions
// @Accept json
// @Produce json
// @Param count query int true "Количество транзакций"
// @Success 200 {array} models.Transaction
// @Failure 400 {object} map[string]interface{} "'error': 'message'"
// @Failure 500 {object}  map[string]interface{} "'error': 'message'"
// @Router /api/transactions [get]
func GetLast(c *fiber.Ctx) error {
	transactions := []*models.Transaction{}

	countStr := c.Query("count")
	countInt, err := strconv.Atoi(countStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "параметр должен быть числом"})
	}

	err = postgres.GetListOfLastTransactions(context.Background(), countInt, &transactions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(transactions)
}

// GetBalance godoc
// @Summary Получить баланс кошелька
// @Description Возвращает текущий баланс по адресу кошелька
// @Tags Wallet
// @Accept json
// @Produce json
// @Param address path string true "Адрес кошелька"
// @Success 200 {number} float64
// @Failure 400 {object} map[string]interface{} "'error': 'message'"
// @Failure 500 {object}  map[string]interface{} "'error': 'message'"
// @Router /api/wallet/{address}/balance [get]
func GetBalance(c *fiber.Ctx) error {
	addr := c.Params("address")
	if addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "адрес кошелька не может быть пустым"})
	}

	// получение баланса
	balance, err := postgres.GetBalanceOfWallet(context.Background(), addr)
	if err != nil {
		if errors.Is(err, postgres.ErrWalletDoesNotExist) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный адрес кошелька"})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(balance)
}
