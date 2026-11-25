package handlers

import (
	"context"
	"strconv"
	"time"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportBudgetHandler struct {
	DB *pgxpool.Pool
}

func (h *ReportBudgetHandler) Summary(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	now := time.Now()
	month, err := strconv.Atoi(c.Query("month", strconv.Itoa(int(now.Month()))))
	if err != nil || month < 1 || month > 12 {
		month = int(now.Month())
	}
	year, err := strconv.Atoi(c.Query("year", strconv.Itoa(now.Year())))
	if err != nil || year < 1900 {
		year = now.Year()
	}

	ctx := context.Background()

	var totalIncome, totalExpense float64
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0)
		 FROM transactions
		 WHERE user_id = $1
		   AND EXTRACT(MONTH FROM transaction_date) = $2
		   AND EXTRACT(YEAR FROM transaction_date) = $3`,
		userID, month, year,
	).Scan(&totalIncome, &totalExpense)

	// Expense breakdown by category
	catRows, err := h.DB.Query(ctx,
		`SELECT COALESCE(c.id, 0), COALESCE(c.name, 'Tanpa Kategori'),
		        COALESCE(c.icon, 'tag'), COALESCE(c.color_hex, '#64748B'),
		        SUM(t.amount) AS total
		 FROM transactions t
		 LEFT JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id
		 WHERE t.user_id = $1
		   AND t.type = 'EXPENSE'
		   AND EXTRACT(MONTH FROM t.transaction_date) = $2
		   AND EXTRACT(YEAR FROM t.transaction_date) = $3
		 GROUP BY c.id, c.name, c.icon, c.color_hex
		 ORDER BY total DESC`,
		userID, month, year,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch category breakdown"})
	}
	defer catRows.Close()

	type categoryStat struct {
		CategoryID int     `json:"category_id"`
		Name       string  `json:"name"`
		Icon       string  `json:"icon"`
		ColorHex   string  `json:"color_hex"`
		Total      float64 `json:"total"`
		Percentage float64 `json:"percentage"`
	}
	byCategory := make([]categoryStat, 0)
	for catRows.Next() {
		var item categoryStat
		if err := catRows.Scan(&item.CategoryID, &item.Name, &item.Icon, &item.ColorHex, &item.Total); err != nil {
			continue
		}
		if totalExpense > 0 {
			item.Percentage = (item.Total / totalExpense) * 100
		}
		byCategory = append(byCategory, item)
	}

	// 6-month cashflow trend ending on (year, month)
	trendRows, err := h.DB.Query(ctx,
		`SELECT EXTRACT(YEAR FROM transaction_date)::INT AS y,
		        EXTRACT(MONTH FROM transaction_date)::INT AS m,
		        COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS income,
		        COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS expense
		 FROM transactions
		 WHERE user_id = $1
		   AND transaction_date >= (MAKE_DATE($2, $3, 1) - INTERVAL '5 months')
		   AND transaction_date < (MAKE_DATE($2, $3, 1) + INTERVAL '1 month')
		 GROUP BY y, m
		 ORDER BY y ASC, m ASC`,
		userID, year, month,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch monthly trend"})
	}
	defer trendRows.Close()

	type trendPoint struct {
		Year    int     `json:"year"`
		Month   int     `json:"month"`
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}
	trend := make([]trendPoint, 0)
	for trendRows.Next() {
		var pt trendPoint
		if err := trendRows.Scan(&pt.Year, &pt.Month, &pt.Income, &pt.Expense); err == nil {
			trend = append(trend, pt)
		}
	}

	return c.JSON(fiber.Map{
		"month": month,
		"year":  year,
		"summary": fiber.Map{
			"income":  totalIncome,
			"expense": totalExpense,
			"total":   totalIncome - totalExpense,
		},
		"expense_by_category": byCategory,
		"monthly_trend":       trend,
	})
}

func (h *ReportBudgetHandler) ListBudgets(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	now := time.Now()
	month, err := strconv.Atoi(c.Query("month", strconv.Itoa(int(now.Month()))))
	if err != nil || month < 1 || month > 12 {
		month = int(now.Month())
	}
	year, err := strconv.Atoi(c.Query("year", strconv.Itoa(now.Year())))
	if err != nil || year < 1900 {
		year = now.Year()
	}

	rows, err := h.DB.Query(context.Background(),
		`SELECT b.id, b.user_id, b.category_id, c.name, c.icon, c.color_hex,
		        b.amount_limit,
		        COALESCE((
		            SELECT SUM(t.amount)
		            FROM transactions t
		            WHERE t.user_id = b.user_id
		              AND t.category_id = b.category_id
		              AND t.type = 'EXPENSE'
		              AND EXTRACT(MONTH FROM t.transaction_date) = b.period_month
		              AND EXTRACT(YEAR FROM t.transaction_date) = b.period_year
		        ), 0) AS spent_amount,
		        b.period_month, b.period_year
		 FROM budgets b
		 JOIN categories c ON c.id = b.category_id AND c.user_id = b.user_id
		 WHERE b.user_id = $1 AND b.period_month = $2 AND b.period_year = $3
		 ORDER BY c.name ASC`,
		userID, month, year,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch budgets"})
	}
	defer rows.Close()

	budgets := make([]models.Budget, 0)
	for rows.Next() {
		var b models.Budget
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.CategoryID, &b.CategoryName, &b.CategoryIcon, &b.CategoryColor,
			&b.AmountLimit, &b.SpentAmount, &b.PeriodMonth, &b.PeriodYear,
		); err == nil {
			budgets = append(budgets, b)
		}
	}

	return c.JSON(fiber.Map{"budgets": budgets})
}

func (h *ReportBudgetHandler) UpsertBudget(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		CategoryID  int     `json:"category_id"`
		AmountLimit float64 `json:"amount_limit"`
		PeriodMonth int     `json:"period_month"`
		PeriodYear  int     `json:"period_year"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if req.CategoryID <= 0 || req.AmountLimit <= 0 || req.PeriodMonth < 1 || req.PeriodMonth > 12 || req.PeriodYear < 1900 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Valid category_id, positive amount_limit, period_month, and period_year are required"})
	}

	ctx := context.Background()
	var catExists bool
	_ = h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`, req.CategoryID, userID).Scan(&catExists)
	if !catExists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	var id int
	err := h.DB.QueryRow(ctx,
		`INSERT INTO budgets (user_id, category_id, amount_limit, period_month, period_year)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, category_id, period_month, period_year)
		 DO UPDATE SET amount_limit = EXCLUDED.amount_limit
		 RETURNING id`,
		userID, req.CategoryID, req.AmountLimit, req.PeriodMonth, req.PeriodYear,
	).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save budget"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "saved": true})
}

func (h *ReportBudgetHandler) DeleteBudget(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid budget ID"})
	}

	tag, err := h.DB.Exec(context.Background(),
		`DELETE FROM budgets WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Budget not found"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
