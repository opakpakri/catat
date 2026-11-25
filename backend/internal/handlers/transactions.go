package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	DB *pgxpool.Pool
}

type txReq struct {
	WalletID        int     `json:"wallet_id"`
	TargetWalletID  *int    `json:"target_wallet_id"`
	CategoryID      *int    `json:"category_id"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Note            string  `json:"note"`
}

func parseTransactionDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Now(), nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid transaction_date format")
}

func applyBalanceDelta(ctx context.Context, tx pgx.Tx, userID, walletID int, targetWalletID *int, txType string, amount float64, reverse bool) error {
	sign := 1.0
	if reverse {
		sign = -1.0
	}

	switch txType {
	case "INCOME":
		tag, err := tx.Exec(ctx,
			`UPDATE wallets SET balance = balance + $1 WHERE id = $2 AND user_id = $3`,
			sign*amount, walletID, userID,
		)
		if err != nil || tag.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "Wallet not found")
		}
	case "EXPENSE":
		tag, err := tx.Exec(ctx,
			`UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND user_id = $3`,
			sign*amount, walletID, userID,
		)
		if err != nil || tag.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "Wallet not found")
		}
	case "TRANSFER":
		if targetWalletID == nil || *targetWalletID == walletID {
			return fiber.NewError(fiber.StatusBadRequest, "Valid target_wallet_id different from wallet_id is required for TRANSFER")
		}
		tag1, err := tx.Exec(ctx,
			`UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND user_id = $3`,
			sign*amount, walletID, userID,
		)
		if err != nil || tag1.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "Source wallet not found")
		}
		tag2, err := tx.Exec(ctx,
			`UPDATE wallets SET balance = balance + $1 WHERE id = $2 AND user_id = $3`,
			sign*amount, *targetWalletID, userID,
		)
		if err != nil || tag2.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "Target wallet not found")
		}
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid transaction type")
	}
	return nil
}

func (h *TransactionHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	query := `
		SELECT t.id, t.user_id, t.wallet_id, w.name,
		       t.target_wallet_id, COALESCE(tw.name, ''),
		       t.category_id, COALESCE(cat.name, ''), COALESCE(cat.icon, 'tag'), COALESCE(cat.color_hex, '#64748B'),
		       t.type, t.amount, t.transaction_date, COALESCE(t.note, ''), t.created_at
		FROM transactions t
		JOIN wallets w ON w.id = t.wallet_id AND w.user_id = t.user_id
		LEFT JOIN wallets tw ON tw.id = t.target_wallet_id AND tw.user_id = t.user_id
		LEFT JOIN categories cat ON cat.id = t.category_id AND cat.user_id = t.user_id
		WHERE t.user_id = $1`

	args := []interface{}{userID}
	argIdx := 2

	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			query += fmt.Sprintf(" AND EXTRACT(MONTH FROM t.transaction_date) = $%d", argIdx)
			args = append(args, m)
			argIdx++
		}
	}
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 1900 {
			query += fmt.Sprintf(" AND EXTRACT(YEAR FROM t.transaction_date) = $%d", argIdx)
			args = append(args, y)
			argIdx++
		}
	}
	if walletStr := c.Query("wallet_id"); walletStr != "" {
		if wid, err := strconv.Atoi(walletStr); err == nil {
			query += fmt.Sprintf(" AND (t.wallet_id = $%d OR t.target_wallet_id = $%d)", argIdx, argIdx)
			args = append(args, wid)
			argIdx++
		}
	}
	if catStr := c.Query("category_id"); catStr != "" {
		if cid, err := strconv.Atoi(catStr); err == nil {
			query += fmt.Sprintf(" AND t.category_id = $%d", argIdx)
			args = append(args, cid)
			argIdx++
		}
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query += fmt.Sprintf(" AND (LOWER(COALESCE(t.note, '')) LIKE $%d OR LOWER(COALESCE(cat.name, '')) LIKE $%d OR LOWER(w.name) LIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+strings.ToLower(search)+"%")
		argIdx++
	}

	query += " ORDER BY t.transaction_date DESC, t.id DESC"

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}
	defer rows.Close()

	list := make([]models.Transaction, 0)
	var totalIncome, totalExpense float64

	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.WalletID, &t.WalletName,
			&t.TargetWalletID, &t.TargetWalletName,
			&t.CategoryID, &t.CategoryName, &t.CategoryIcon, &t.CategoryColor,
			&t.Type, &t.Amount, &t.TransactionDate, &t.Note, &t.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan transaction"})
		}
		if t.Type == "INCOME" {
			totalIncome += t.Amount
		} else if t.Type == "EXPENSE" {
			totalExpense += t.Amount
		}
		list = append(list, t)
	}

	return c.JSON(fiber.Map{
		"transactions": list,
		"summary": fiber.Map{
			"income":  totalIncome,
			"expense": totalExpense,
			"total":   totalIncome - totalExpense,
		},
	})
}

func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req txReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if req.Amount <= 0 || req.WalletID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "wallet_id and positive amount are required"})
	}

	txDate, err := parseTransactionDate(req.TransactionDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	// Cross-tenant validation: verify category belongs to user if provided
	if req.CategoryID != nil && req.Type != "TRANSFER" {
		var exists bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`, *req.CategoryID, userID).Scan(&exists)
		if err != nil || !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
		}
	}
	if req.Type == "TRANSFER" {
		req.CategoryID = nil
	}

	// Atomic balance update + wallet ownership check
	if err := applyBalanceDelta(ctx, tx, userID, req.WalletID, req.TargetWalletID, req.Type, req.Amount, false); err != nil {
		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update wallet balance"})
	}

	var created models.Transaction
	err = tx.QueryRow(ctx,
		`INSERT INTO transactions (user_id, wallet_id, target_wallet_id, category_id, type, amount, transaction_date, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_id, wallet_id, target_wallet_id, category_id, type, amount, transaction_date, COALESCE(note, ''), created_at`,
		userID, req.WalletID, req.TargetWalletID, req.CategoryID, req.Type, req.Amount, txDate, req.Note,
	).Scan(
		&created.ID, &created.UserID, &created.WalletID, &created.TargetWalletID,
		&created.CategoryID, &created.Type, &created.Amount, &created.TransactionDate, &created.Note, &created.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"transaction": created})
}

func (h *TransactionHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	txID, err := strconv.Atoi(c.Params("id"))
	if err != nil || txID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction ID"})
	}

	var req txReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if req.Amount <= 0 || req.WalletID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "wallet_id and positive amount are required"})
	}

	txDate, err := parseTransactionDate(req.TransactionDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	// Fetch existing transaction scoped to user_id
	var oldWalletID int
	var oldTargetWalletID *int
	var oldType string
	var oldAmount float64
	err = tx.QueryRow(ctx,
		`SELECT wallet_id, target_wallet_id, type, amount FROM transactions WHERE id = $1 AND user_id = $2`,
		txID, userID,
	).Scan(&oldWalletID, &oldTargetWalletID, &oldType, &oldAmount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Reverse old balance effect
	if err := applyBalanceDelta(ctx, tx, userID, oldWalletID, oldTargetWalletID, oldType, oldAmount, true); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reverse previous balance"})
	}

	// Validate new category if present
	if req.CategoryID != nil && req.Type != "TRANSFER" {
		var exists bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`, *req.CategoryID, userID).Scan(&exists)
		if err != nil || !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
		}
	}
	if req.Type == "TRANSFER" {
		req.CategoryID = nil
	}

	// Apply new balance effect
	if err := applyBalanceDelta(ctx, tx, userID, req.WalletID, req.TargetWalletID, req.Type, req.Amount, false); err != nil {
		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update wallet balance"})
	}

	var updated models.Transaction
	err = tx.QueryRow(ctx,
		`UPDATE transactions
		 SET wallet_id = $1, target_wallet_id = $2, category_id = $3, type = $4, amount = $5, transaction_date = $6, note = $7
		 WHERE id = $8 AND user_id = $9
		 RETURNING id, user_id, wallet_id, target_wallet_id, category_id, type, amount, transaction_date, COALESCE(note, ''), created_at`,
		req.WalletID, req.TargetWalletID, req.CategoryID, req.Type, req.Amount, txDate, req.Note, txID, userID,
	).Scan(
		&updated.ID, &updated.UserID, &updated.WalletID, &updated.TargetWalletID,
		&updated.CategoryID, &updated.Type, &updated.Amount, &updated.TransactionDate, &updated.Note, &updated.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update transaction"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit update"})
	}

	return c.JSON(fiber.Map{"transaction": updated})
}

func (h *TransactionHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	txID, err := strconv.Atoi(c.Params("id"))
	if err != nil || txID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction ID"})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	var walletID int
	var targetWalletID *int
	var txType string
	var amount float64
	err = tx.QueryRow(ctx,
		`SELECT wallet_id, target_wallet_id, type, amount FROM transactions WHERE id = $1 AND user_id = $2`,
		txID, userID,
	).Scan(&walletID, &targetWalletID, &txType, &amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if err := applyBalanceDelta(ctx, tx, userID, walletID, targetWalletID, txType, amount, true); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to restore wallet balance"})
	}

	tag, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, txID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit delete"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
