package routes

import (
	"catat-api/internal/config"
	"catat-api/internal/handlers"
	"catat-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool, cfg *config.Config) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": "CATAT API"})
	})

	authH := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	txH := &handlers.TransactionHandler{DB: db}
	calH := &handlers.CalendarHandler{DB: db}
	repH := &handlers.ReportBudgetHandler{DB: db}
	accH := &handlers.AccountDebtHandler{DB: db}
	moreH := &handlers.MoreHandler{DB: db}

	api := app.Group("/api")

	// Public Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authH.Register)
	auth.Post("/login", authH.Login)

	// Protected routes (JWT + Tenant Context Middleware)
	protected := api.Group("", middleware.AuthTenant(cfg.JWTSecret))

	protected.Get("/auth/me", authH.Me)

	// Users management (Role: ADMIN only)
	users := protected.Group("/users", middleware.AdminOnly(db))
	users.Get("/", authH.ListUsers)
	users.Post("/", authH.CreateUser)
	users.Put("/:id", authH.UpdateUser)
	users.Delete("/:id", authH.DeleteUser)

	// Transactions (Daily)
	protected.Get("/transactions", txH.List)
	protected.Post("/transactions", txH.Create)
	protected.Put("/transactions/:id", txH.Update)
	protected.Delete("/transactions/:id", txH.Delete)

	// Calendar & Events (Monthly)
	protected.Get("/calendar/monthly", calH.Monthly)
	protected.Post("/events", calH.CreateEvent)
	protected.Put("/events/:id", calH.UpdateEvent)
	protected.Patch("/events/:id/toggle", calH.ToggleEvent)
	protected.Delete("/events/:id", calH.DeleteEvent)

	// Reports & Budgets (Stats)
	protected.Get("/reports/summary", repH.Summary)
	protected.Get("/budgets", repH.ListBudgets)
	protected.Post("/budgets", repH.UpsertBudget)
	protected.Delete("/budgets/:id", repH.DeleteBudget)

	// Wallets & Debts (Accounts)
	protected.Get("/wallets", accH.ListWallets)
	protected.Post("/wallets", accH.CreateWallet)
	protected.Put("/wallets/:id", accH.UpdateWallet)
	protected.Get("/debts", accH.ListDebts)
	protected.Post("/debts", accH.CreateDebt)
	protected.Patch("/debts/:id/pay", accH.PayDebt)

	// Categories & Export (More)
	protected.Get("/categories", moreH.ListCategories)
	protected.Post("/categories", moreH.CreateCategory)
	protected.Delete("/categories/:id", moreH.DeleteCategory)
	protected.Get("/export/csv", moreH.ExportCSV)
	protected.Get("/backup/json", moreH.BackupJSON)
	protected.Post("/backup/json", moreH.RestoreJSON)
}
