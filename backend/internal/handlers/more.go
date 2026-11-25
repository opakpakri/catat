package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MoreHandler struct {
	DB *pgxpool.Pool
}

func (h *MoreHandler) ListCategories(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	query := `SELECT id, user_id, name, type, icon, color_hex, created_at FROM categories WHERE user_id = $1`
	args := []interface{}{userID}
	if t := strings.ToUpper(strings.TrimSpace(c.Query("type"))); t == "INCOME" || t == "EXPENSE" {
		query += " AND type = $2"
		args = append(args, t)
	}
	query += " ORDER BY type ASC, name ASC"

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	defer rows.Close()

	cats := make([]models.Category, 0)
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.ColorHex, &cat.CreatedAt); err == nil {
			cats = append(cats, cat)
		}
	}

	return c.JSON(fiber.Map{"categories": cats})
}

func (h *MoreHandler) CreateCategory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Icon     string `json:"icon"`
		ColorHex string `json:"color_hex"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if req.Name == "" || (req.Type != "INCOME" && req.Type != "EXPENSE") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and type (INCOME/EXPENSE) are required"})
	}
	if req.Icon == "" {
		req.Icon = "tag"
	}
	if req.ColorHex == "" {
		req.ColorHex = "#64748B"
	}

	var cat models.Category
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO categories (user_id, name, type, icon, color_hex)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, type, icon, color_hex, created_at`,
		userID, req.Name, req.Type, req.Icon, req.ColorHex,
	).Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.ColorHex, &cat.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Category already exists"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to create category"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"category": cat})
}

func (h *MoreHandler) DeleteCategory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid category ID"})
	}

	tag, err := h.DB.Exec(context.Background(),
		`DELETE FROM categories WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}

func (h *MoreHandler) ExportCSV(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rows, err := h.DB.Query(context.Background(),
		`SELECT TO_CHAR(t.transaction_date, 'YYYY-MM-DD HH24:MI'), t.type,
		        w.name, COALESCE(tw.name, ''), COALESCE(c.name, ''), t.amount, COALESCE(t.note, '')
		 FROM transactions t
		 JOIN wallets w ON w.id = t.wallet_id AND w.user_id = t.user_id
		 LEFT JOIN wallets tw ON tw.id = t.target_wallet_id AND tw.user_id = t.user_id
		 LEFT JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id
		 WHERE t.user_id = $1
		 ORDER BY t.transaction_date DESC`,
		userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to export transactions"})
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"Tanggal", "Tipe", "Dompet", "Dompet Tujuan", "Kategori", "Nominal", "Catatan"})

	for rows.Next() {
		var dateStr, txType, wallet, targetWallet, category, note string
		var amount float64
		if err := rows.Scan(&dateStr, &txType, &wallet, &targetWallet, &category, &amount, &note); err == nil {
			_ = writer.Write([]string{dateStr, txType, wallet, targetWallet, category, fmt.Sprintf("%.2f", amount), note})
		}
	}
	writer.Flush()

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="catat_transaksi.csv"`)
	return c.Send(buf.Bytes())
}

func (h *MoreHandler) BackupJSON(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	// Export wallets, categories, transactions, events, debts
	wallets := make([]models.Wallet, 0)
	if rows, err := h.DB.Query(ctx, `SELECT id, user_id, name, type, balance, color_hex, created_at FROM wallets WHERE user_id = $1`, userID); err == nil {
		for rows.Next() {
			var w models.Wallet
			if rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Type, &w.Balance, &w.ColorHex, &w.CreatedAt) == nil {
				wallets = append(wallets, w)
			}
		}
		rows.Close()
	}

	categories := make([]models.Category, 0)
	if rows, err := h.DB.Query(ctx, `SELECT id, user_id, name, type, icon, color_hex, created_at FROM categories WHERE user_id = $1`, userID); err == nil {
		for rows.Next() {
			var cat models.Category
			if rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.ColorHex, &cat.CreatedAt) == nil {
				categories = append(categories, cat)
			}
		}
		rows.Close()
	}

	transactions := make([]models.Transaction, 0)
	if rows, err := h.DB.Query(ctx, `SELECT id, user_id, wallet_id, target_wallet_id, category_id, type, amount, transaction_date, COALESCE(note, ''), created_at FROM transactions WHERE user_id = $1`, userID); err == nil {
		for rows.Next() {
			var t models.Transaction
			if rows.Scan(&t.ID, &t.UserID, &t.WalletID, &t.TargetWalletID, &t.CategoryID, &t.Type, &t.Amount, &t.TransactionDate, &t.Note, &t.CreatedAt) == nil {
				transactions = append(transactions, t)
			}
		}
		rows.Close()
	}

	events := make([]models.CalendarEvent, 0)
	if rows, err := h.DB.Query(ctx, `SELECT id, user_id, debt_id, title, COALESCE(description, ''), TO_CHAR(event_date, 'YYYY-MM-DD'), CASE WHEN event_time IS NOT NULL THEN TO_CHAR(event_time, 'HH24:MI') ELSE NULL END, is_completed, color_hex, created_at FROM calendar_events WHERE user_id = $1`, userID); err == nil {
		for rows.Next() {
			var ev models.CalendarEvent
			if rows.Scan(&ev.ID, &ev.UserID, &ev.DebtID, &ev.Title, &ev.Description, &ev.EventDate, &ev.EventTime, &ev.IsCompleted, &ev.ColorHex, &ev.CreatedAt) == nil {
				events = append(events, ev)
			}
		}
		rows.Close()
	}

	debts := make([]models.Debt, 0)
	if rows, err := h.DB.Query(ctx, `SELECT id, user_id, person_name, type, total_amount, paid_amount, CASE WHEN due_date IS NOT NULL THEN TO_CHAR(due_date, 'YYYY-MM-DD') ELSE NULL END, status, COALESCE(note, ''), created_at FROM debts WHERE user_id = $1`, userID); err == nil {
		for rows.Next() {
			var d models.Debt
			if rows.Scan(&d.ID, &d.UserID, &d.PersonName, &d.Type, &d.TotalAmount, &d.PaidAmount, &d.DueDate, &d.Status, &d.Note, &d.CreatedAt) == nil {
				debts = append(debts, d)
			}
		}
		rows.Close()
	}

	c.Set("Content-Disposition", `attachment; filename="catat_backup.json"`)
	return c.JSON(fiber.Map{
		"version":      "1.0",
		"wallets":      wallets,
		"categories":   categories,
		"transactions": transactions,
		"events":       events,
		"debts":        debts,
	})
}

func (h *MoreHandler) RestoreJSON(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var payload struct {
		Wallets    []models.Wallet        `json:"wallets"`
		Categories []models.Category      `json:"categories"`
		Events     []models.CalendarEvent `json:"events"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid backup JSON"})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	restoredWallets := 0
	for _, w := range payload.Wallets {
		tag, err := tx.Exec(ctx,
			`INSERT INTO wallets (user_id, name, type, balance, color_hex)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (user_id, name) DO NOTHING`,
			userID, w.Name, w.Type, w.Balance, w.ColorHex,
		)
		if err == nil && tag.RowsAffected() > 0 {
			restoredWallets++
		}
	}

	restoredCats := 0
	for _, cat := range payload.Categories {
		tag, err := tx.Exec(ctx,
			`INSERT INTO categories (user_id, name, type, icon, color_hex)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (user_id, name, type) DO NOTHING`,
			userID, cat.Name, cat.Type, cat.Icon, cat.ColorHex,
		)
		if err == nil && tag.RowsAffected() > 0 {
			restoredCats++
		}
	}

	restoredEvents := 0
	for _, ev := range payload.Events {
		tag, err := tx.Exec(ctx,
			`INSERT INTO calendar_events (user_id, title, description, event_date, event_time, is_completed, color_hex)
			 VALUES ($1, $2, $3, $4::DATE, $5::TIME, $6, $7)`,
			userID, ev.Title, ev.Description, ev.EventDate, ev.EventTime, ev.IsCompleted, ev.ColorHex,
		)
		if err == nil && tag.RowsAffected() > 0 {
			restoredEvents++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to restore backup"})
	}

	return c.JSON(fiber.Map{
		"restored": true,
		"counts": fiber.Map{
			"wallets":    restoredWallets,
			"categories": restoredCats,
			"events":     restoredEvents,
		},
	})
}
