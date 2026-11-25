package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CalendarHandler struct {
	DB *pgxpool.Pool
}

type eventReq struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	EventDate   string  `json:"event_date"`
	EventTime   *string `json:"event_time"`
	ColorHex    string  `json:"color_hex"`
}

func (h *CalendarHandler) Monthly(c *fiber.Ctx) error {
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

	// Daily aggregation for the month
	dailyRows, err := h.DB.Query(ctx,
		`SELECT TO_CHAR(transaction_date, 'YYYY-MM-DD') AS day,
		        COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS income,
		        COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS expense
		 FROM transactions
		 WHERE user_id = $1
		   AND EXTRACT(MONTH FROM transaction_date) = $2
		   AND EXTRACT(YEAR FROM transaction_date) = $3
		 GROUP BY day
		 ORDER BY day`,
		userID, month, year,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch monthly aggregation"})
	}
	defer dailyRows.Close()

	type daySummary struct {
		Date    string  `json:"date"`
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}
	days := make([]daySummary, 0)
	var totalIncome, totalExpense float64
	for dailyRows.Next() {
		var d daySummary
		if err := dailyRows.Scan(&d.Date, &d.Income, &d.Expense); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan daily summary"})
		}
		totalIncome += d.Income
		totalExpense += d.Expense
		days = append(days, d)
	}

	// Events for the month
	eventRows, err := h.DB.Query(ctx,
		`SELECT id, user_id, debt_id, title, COALESCE(description, ''),
		        TO_CHAR(event_date, 'YYYY-MM-DD'),
		        CASE WHEN event_time IS NOT NULL THEN TO_CHAR(event_time, 'HH24:MI') ELSE NULL END,
		        is_completed, color_hex, created_at
		 FROM calendar_events
		 WHERE user_id = $1
		   AND EXTRACT(MONTH FROM event_date) = $2
		   AND EXTRACT(YEAR FROM event_date) = $3
		 ORDER BY event_date ASC, event_time ASC NULLS LAST, id ASC`,
		userID, month, year,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch calendar events"})
	}
	defer eventRows.Close()

	events := make([]models.CalendarEvent, 0)
	for eventRows.Next() {
		var ev models.CalendarEvent
		if err := eventRows.Scan(
			&ev.ID, &ev.UserID, &ev.DebtID, &ev.Title, &ev.Description,
			&ev.EventDate, &ev.EventTime, &ev.IsCompleted, &ev.ColorHex, &ev.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan event"})
		}
		events = append(events, ev)
	}

	return c.JSON(fiber.Map{
		"month": month,
		"year":  year,
		"summary": fiber.Map{
			"income":  totalIncome,
			"expense": totalExpense,
			"total":   totalIncome - totalExpense,
		},
		"days":   days,
		"events": events,
	})
}

func (h *CalendarHandler) CreateEvent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req eventReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.EventDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and event_date (YYYY-MM-DD) are required"})
	}
	if req.ColorHex == "" {
		req.ColorHex = "#10B981"
	}
	if req.EventTime != nil && strings.TrimSpace(*req.EventTime) == "" {
		req.EventTime = nil
	}

	var ev models.CalendarEvent
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO calendar_events (user_id, title, description, event_date, event_time, color_hex)
		 VALUES ($1, $2, $3, $4::DATE, $5::TIME, $6)
		 RETURNING id, user_id, debt_id, title, COALESCE(description, ''),
		           TO_CHAR(event_date, 'YYYY-MM-DD'),
		           CASE WHEN event_time IS NOT NULL THEN TO_CHAR(event_time, 'HH24:MI') ELSE NULL END,
		           is_completed, color_hex, created_at`,
		userID, req.Title, req.Description, req.EventDate, req.EventTime, req.ColorHex,
	).Scan(
		&ev.ID, &ev.UserID, &ev.DebtID, &ev.Title, &ev.Description,
		&ev.EventDate, &ev.EventTime, &ev.IsCompleted, &ev.ColorHex, &ev.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to create event"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"event": ev})
}

func (h *CalendarHandler) UpdateEvent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	var req eventReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.EventDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and event_date are required"})
	}
	if req.ColorHex == "" {
		req.ColorHex = "#10B981"
	}
	if req.EventTime != nil && strings.TrimSpace(*req.EventTime) == "" {
		req.EventTime = nil
	}

	var ev models.CalendarEvent
	err = h.DB.QueryRow(context.Background(),
		`UPDATE calendar_events
		 SET title = $1, description = $2, event_date = $3::DATE, event_time = $4::TIME, color_hex = $5
		 WHERE id = $6 AND user_id = $7
		 RETURNING id, user_id, debt_id, title, COALESCE(description, ''),
		           TO_CHAR(event_date, 'YYYY-MM-DD'),
		           CASE WHEN event_time IS NOT NULL THEN TO_CHAR(event_time, 'HH24:MI') ELSE NULL END,
		           is_completed, color_hex, created_at`,
		req.Title, req.Description, req.EventDate, req.EventTime, req.ColorHex, id, userID,
	).Scan(
		&ev.ID, &ev.UserID, &ev.DebtID, &ev.Title, &ev.Description,
		&ev.EventDate, &ev.EventTime, &ev.IsCompleted, &ev.ColorHex, &ev.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	return c.JSON(fiber.Map{"event": ev})
}

func (h *CalendarHandler) ToggleEvent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	var currentCompleted bool
	err = h.DB.QueryRow(context.Background(),
		`SELECT is_completed FROM calendar_events WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&currentCompleted)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}
	if currentCompleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Acara yang sudah selesai tidak dapat diubah lagi"})
	}

	var isCompleted bool
	err = h.DB.QueryRow(context.Background(),
		`UPDATE calendar_events SET is_completed = TRUE
		 WHERE id = $1 AND user_id = $2
		 RETURNING is_completed`,
		id, userID,
	).Scan(&isCompleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update event"})
	}

	return c.JSON(fiber.Map{"id": id, "is_completed": isCompleted})
}

func (h *CalendarHandler) DeleteEvent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	tag, err := h.DB.Exec(context.Background(),
		`DELETE FROM calendar_events WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
