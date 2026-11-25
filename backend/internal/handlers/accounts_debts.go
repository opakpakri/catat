package handlers

import (
	"context"
	"strconv"
	"strings"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountDebtHandler struct {
	DB *pgxpool.Pool
}

func (h *AccountDebtHandler) ListWallets(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rows, err := h.DB.Query(context.Background(),
		`SELECT id, user_id, name, type, balance, color_hex, created_at
		 FROM wallets
		 WHERE user_id = $1
		 ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch wallets"})
	}
	defer rows.Close()

	wallets := make([]models.Wallet, 0)
	var netWorth float64
	for rows.Next() {
		var w models.Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Type, &w.Balance, &w.ColorHex, &w.CreatedAt); err == nil {
			netWorth += w.Balance
			wallets = append(wallets, w)
		}
	}

	return c.JSON(fiber.Map{
		"wallets":   wallets,
		"net_worth": netWorth,
	})
}

func (h *AccountDebtHandler) CreateWallet(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		Name     string  `json:"name"`
		Type     string  `json:"type"`
		Balance  float64 `json:"balance"`
		ColorHex string  `json:"color_hex"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if req.Type == "" {
		req.Type = "CASH"
	}
	if req.ColorHex == "" {
		req.ColorHex = "#3B82F6"
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Wallet name is required"})
	}

	var w models.Wallet
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO wallets (user_id, name, type, balance, color_hex)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, type, balance, color_hex, created_at`,
		userID, req.Name, req.Type, req.Balance, req.ColorHex,
	).Scan(&w.ID, &w.UserID, &w.Name, &w.Type, &w.Balance, &w.ColorHex, &w.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Wallet name already exists"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to create wallet"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"wallet": w})
}

func (h *AccountDebtHandler) UpdateWallet(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid wallet ID"})
	}

	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		ColorHex string `json:"color_hex"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Wallet name is required"})
	}
	if req.Type == "" {
		req.Type = "CASH"
	}
	if req.ColorHex == "" {
		req.ColorHex = "#3B82F6"
	}

	var w models.Wallet
	err = h.DB.QueryRow(context.Background(),
		`UPDATE wallets SET name = $1, type = $2, color_hex = $3
		 WHERE id = $4 AND user_id = $5
		 RETURNING id, user_id, name, type, balance, color_hex, created_at`,
		req.Name, req.Type, req.ColorHex, id, userID,
	).Scan(&w.ID, &w.UserID, &w.Name, &w.Type, &w.Balance, &w.ColorHex, &w.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
	}

	return c.JSON(fiber.Map{"wallet": w})
}

func (h *AccountDebtHandler) ListDebts(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	query := `
		SELECT d.id, d.user_id, d.wallet_id, COALESCE(w.name, ''),
		       d.person_name, d.type, d.total_amount, d.paid_amount,
		       CASE WHEN d.due_date IS NOT NULL THEN TO_CHAR(d.due_date, 'YYYY-MM-DD') ELSE NULL END,
		       d.status, COALESCE(d.note, ''), d.created_at
		FROM debts d
		LEFT JOIN wallets w ON w.id = d.wallet_id AND w.user_id = d.user_id
		WHERE d.user_id = $1`
	args := []interface{}{userID}
	idx := 2

	if t := strings.ToUpper(strings.TrimSpace(c.Query("type"))); t == "DEBT" || t == "RECEIVABLE" {
		query += " AND d.type = $" + strconv.Itoa(idx)
		args = append(args, t)
		idx++
	}
	if s := strings.ToUpper(strings.TrimSpace(c.Query("status"))); s == "UNPAID" || s == "PARTIAL" || s == "PAID" {
		query += " AND d.status = $" + strconv.Itoa(idx)
		args = append(args, s)
		idx++
	}
	query += " ORDER BY CASE WHEN d.status = 'PAID' THEN 1 ELSE 0 END ASC, d.due_date ASC NULLS LAST, d.id DESC"

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch debts"})
	}
	defer rows.Close()

	debts := make([]models.Debt, 0)
	for rows.Next() {
		var d models.Debt
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.WalletID, &d.WalletName,
			&d.PersonName, &d.Type, &d.TotalAmount,
			&d.PaidAmount, &d.DueDate, &d.Status, &d.Note, &d.CreatedAt,
		); err == nil {
			debts = append(debts, d)
		}
	}

	return c.JSON(fiber.Map{"debts": debts})
}

func (h *AccountDebtHandler) CreateDebt(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		WalletID       int     `json:"wallet_id"`
		PersonName     string  `json:"person_name"`
		Type           string  `json:"type"`
		TotalAmount    float64 `json:"total_amount"`
		DueDate        string  `json:"due_date"`
		Note           string  `json:"note"`
		CreateReminder bool    `json:"create_reminder"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.PersonName = strings.TrimSpace(req.PersonName)
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	req.DueDate = strings.TrimSpace(req.DueDate)
	if req.WalletID <= 0 || req.PersonName == "" || req.TotalAmount <= 0 || req.DueDate == "" || (req.Type != "DEBT" && req.Type != "RECEIVABLE") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "wallet_id, person_name, valid type (DEBT/RECEIVABLE), due_date, and positive total_amount are required",
		})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	// Verify wallet belongs to user and update balance atomically
	// DEBT (saya pinjam) => wallet balance +total_amount
	// RECEIVABLE (dipinjam orang) => wallet balance -total_amount
	delta := req.TotalAmount
	if req.Type == "RECEIVABLE" {
		delta = -req.TotalAmount
	}
	var walletName string
	err = tx.QueryRow(ctx,
		`UPDATE wallets SET balance = balance + $1 WHERE id = $2 AND user_id = $3 RETURNING name`,
		delta, req.WalletID, userID,
	).Scan(&walletName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
	}

	var d models.Debt
	err = tx.QueryRow(ctx,
		`INSERT INTO debts (user_id, wallet_id, person_name, type, total_amount, paid_amount, due_date, status, note)
		 VALUES ($1, $2, $3, $4, $5, 0, $6::DATE, 'UNPAID', $7)
		 RETURNING id, user_id, wallet_id, person_name, type, total_amount, paid_amount,
		           CASE WHEN due_date IS NOT NULL THEN TO_CHAR(due_date, 'YYYY-MM-DD') ELSE NULL END,
		           status, COALESCE(note, ''), created_at`,
		userID, req.WalletID, req.PersonName, req.Type, req.TotalAmount, req.DueDate, req.Note,
	).Scan(
		&d.ID, &d.UserID, &d.WalletID, &d.PersonName, &d.Type, &d.TotalAmount,
		&d.PaidAmount, &d.DueDate, &d.Status, &d.Note, &d.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to create debt record"})
	}
	d.WalletName = walletName

	if req.CreateReminder {
		label := "Jatuh Tempo Utang: " + req.PersonName
		color := "#F59E0B"
		if req.Type == "RECEIVABLE" {
			label = "Tagih Piutang: " + req.PersonName
			color = "#10B981"
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO calendar_events (user_id, debt_id, title, description, event_date, color_hex)
			 VALUES ($1, $2, $3, $4, $5::DATE, $6)`,
			userID, d.ID, label, req.Note, req.DueDate, color,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create calendar reminder"})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit debt"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"debt": d})
}

func (h *AccountDebtHandler) PayDebt(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid debt ID"})
	}

	var req struct {
		Amount   float64 `json:"amount"`
		WalletID *int    `json:"wallet_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Positive payment amount is required"})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	var totalAmount, paidAmount float64
	var debtType string
	var debtWalletID *int
	err = tx.QueryRow(ctx,
		`SELECT total_amount, paid_amount, type, wallet_id FROM debts WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&totalAmount, &paidAmount, &debtType, &debtWalletID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Debt record not found"})
	}

	remaining := totalAmount - paidAmount
	if req.Amount > remaining {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nominal pembayaran tidak boleh melebihi sisa yang harus dilunasi",
		})
	}
	applied := req.Amount

	targetWalletID := debtWalletID
	if req.WalletID != nil && *req.WalletID > 0 {
		targetWalletID = req.WalletID
	}
	if targetWalletID != nil && applied > 0 {
		// Paying DEBT => wallet balance -applied
		// Receiving RECEIVABLE => wallet balance +applied
		delta := applied
		if debtType == "DEBT" {
			delta = -applied
		}
		tag, err := tx.Exec(ctx,
			`UPDATE wallets SET balance = balance + $1 WHERE id = $2 AND user_id = $3`,
			delta, *targetWalletID, userID,
		)
		if err != nil || tag.RowsAffected() == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
		}
	}

	newPaid := paidAmount + applied
	status := "PARTIAL"
	if newPaid >= totalAmount {
		status = "PAID"
	}

	var d models.Debt
	err = tx.QueryRow(ctx,
		`UPDATE debts SET paid_amount = $1, status = $2
		 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, wallet_id, person_name, type, total_amount, paid_amount,
		           CASE WHEN due_date IS NOT NULL THEN TO_CHAR(due_date, 'YYYY-MM-DD') ELSE NULL END,
		           status, COALESCE(note, ''), created_at`,
		newPaid, status, id, userID,
	).Scan(
		&d.ID, &d.UserID, &d.WalletID, &d.PersonName, &d.Type, &d.TotalAmount,
		&d.PaidAmount, &d.DueDate, &d.Status, &d.Note, &d.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment"})
	}

	if status == "PAID" {
		_, _ = tx.Exec(ctx,
			`UPDATE calendar_events SET is_completed = TRUE WHERE debt_id = $1 AND user_id = $2`,
			id, userID,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit payment"})
	}

	return c.JSON(fiber.Map{"debt": d})
}
