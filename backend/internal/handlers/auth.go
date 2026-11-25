package handlers

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"catat-api/internal/middleware"
	"catat-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type pinReq struct {
	PIN     string `json:"pin"`
	Enabled *bool  `json:"enabled"`
}

var pinRegex = regexp.MustCompile(`^\d{6}$`)

func (h *AuthHandler) issueToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}

// SeedDefaultUserResources seeds initial wallets and categories for a newly created user inside tx.
func SeedDefaultUserResources(ctx context.Context, tx pgx.Tx, userID int) error {
	defaultWallets := []struct {
		Name  string
		Type  string
		Color string
	}{
		{"Tunai", "CASH", "#10B981"},
		{"Bank", "BANK", "#3B82F6"},
		{"E-Wallet", "EWALLET", "#8B5CF6"},
	}
	for _, w := range defaultWallets {
		if _, err := tx.Exec(ctx,
			`INSERT INTO wallets (user_id, name, type, balance, color_hex) VALUES ($1, $2, $3, 0, $4) ON CONFLICT (user_id, name) DO NOTHING`,
			userID, w.Name, w.Type, w.Color,
		); err != nil {
			return err
		}
	}

	defaultCategories := []struct {
		Name  string
		Type  string
		Icon  string
		Color string
	}{
		{"Gaji", "INCOME", "briefcase", "#3B82F6"},
		{"Usaha", "INCOME", "store", "#10B981"},
		{"Bonus", "INCOME", "gift", "#06B6D4"},
		{"Lainnya", "INCOME", "plus-circle", "#64748B"},
		{"Makan & Minum", "EXPENSE", "utensils", "#F05A4F"},
		{"Transportasi", "EXPENSE", "car", "#F59E0B"},
		{"Belanja", "EXPENSE", "shopping-bag", "#EC4899"},
		{"Tagihan", "EXPENSE", "receipt", "#EF4444"},
		{"Hiburan", "EXPENSE", "film", "#8B5CF6"},
		{"Kesehatan", "EXPENSE", "heart-pulse", "#14B8A6"},
		{"Lainnya", "EXPENSE", "tag", "#64748B"},
	}
	for _, cat := range defaultCategories {
		if _, err := tx.Exec(ctx,
			`INSERT INTO categories (user_id, name, type, icon, color_hex) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, name, type) DO NOTHING`,
			userID, cat.Name, cat.Type, cat.Icon, cat.Color,
		); err != nil {
			return err
		}
	}
	return nil
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": "Registrasi publik dinonaktifkan. Pembuatan akun hanya dapat dilakukan oleh Admin.",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	var user models.User
	var passwordHash string
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, full_name, email, role, password_hash, is_pin_enabled, currency, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.FullName, &user.Email, &user.Role, &passwordHash, &user.IsPINEnabled, &user.Currency, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	token, err := h.issueToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var user models.User
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, full_name, email, role, is_pin_enabled, currency, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.FullName, &user.Email, &user.Role, &user.IsPINEnabled, &user.Currency, &user.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) SetupPIN(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req pinReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	if req.Enabled != nil && !*req.Enabled {
		_, err := h.DB.Exec(context.Background(),
			`UPDATE users SET security_pin_hash = NULL, is_pin_enabled = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			userID,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to disable PIN"})
		}
		return c.JSON(fiber.Map{"is_pin_enabled": false, "message": "PIN disabled"})
	}

	if !pinRegex.MatchString(req.PIN) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PIN must be exactly 6 digits"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash PIN"})
	}

	_, err = h.DB.Exec(context.Background(),
		`UPDATE users SET security_pin_hash = $1, is_pin_enabled = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		string(hash), userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update PIN"})
	}

	return c.JSON(fiber.Map{"is_pin_enabled": true, "message": "6-digit PIN enabled"})
}

func (h *AuthHandler) VerifyPIN(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req pinReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	if !pinRegex.MatchString(req.PIN) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PIN must be 6 digits"})
	}

	var pinHash *string
	var enabled bool
	err := h.DB.QueryRow(context.Background(),
		`SELECT security_pin_hash, is_pin_enabled FROM users WHERE id = $1`,
		userID,
	).Scan(&pinHash, &enabled)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if !enabled || pinHash == nil {
		return c.JSON(fiber.Map{"verified": true})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*pinHash), []byte(req.PIN)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "PIN salah", "verified": false})
	}

	return c.JSON(fiber.Map{"verified": true})
}

func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	rows, err := h.DB.Query(context.Background(),
		`SELECT id, full_name, email, role, is_pin_enabled, currency, created_at
		 FROM users
		 ORDER BY id DESC`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memuat daftar pengguna"})
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Role, &u.IsPINEnabled, &u.Currency, &u.CreatedAt); err == nil {
			users = append(users, u)
		}
	}

	return c.JSON(fiber.Map{"users": users})
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Currency string `json:"currency"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	if req.Role != "ADMIN" && req.Role != "USER" {
		req.Role = "USER"
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.FullName == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama lengkap dan email wajib diisi"})
	}

	ctx := context.Background()
	var user models.User
	if strings.TrimSpace(req.Password) != "" {
		if len(req.Password) < 6 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password minimal 6 karakter"})
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
		}
		err = h.DB.QueryRow(ctx,
			`UPDATE users SET full_name = $1, email = $2, password_hash = $3, role = $4, currency = $5, updated_at = CURRENT_TIMESTAMP
			 WHERE id = $6
			 RETURNING id, full_name, email, role, is_pin_enabled, currency, created_at`,
			req.FullName, req.Email, string(hash), req.Role, req.Currency, id,
		).Scan(&user.ID, &user.FullName, &user.Email, &user.Role, &user.IsPINEnabled, &user.Currency, &user.CreatedAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal memperbarui pengguna (email mungkin sudah digunakan)"})
		}
	} else {
		err = h.DB.QueryRow(ctx,
			`UPDATE users SET full_name = $1, email = $2, role = $3, currency = $4, updated_at = CURRENT_TIMESTAMP
			 WHERE id = $5
			 RETURNING id, full_name, email, role, is_pin_enabled, currency, created_at`,
			req.FullName, req.Email, req.Role, req.Currency, id,
		).Scan(&user.ID, &user.FullName, &user.Email, &user.Role, &user.IsPINEnabled, &user.Currency, &user.CreatedAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal memperbarui pengguna (email mungkin sudah digunakan)"})
		}
	}

	return c.JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Currency string `json:"currency"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	if req.Role != "ADMIN" && req.Role != "USER" {
		req.Role = "USER"
	}
	if req.FullName == "" || req.Email == "" || len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama lengkap, email valid, dan password minimal 6 karakter wajib diisi",
		})
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(ctx)

	var user models.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (full_name, email, password_hash, role, currency)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, full_name, email, role, is_pin_enabled, currency, created_at`,
		req.FullName, req.Email, string(hash), req.Role, req.Currency,
	).Scan(&user.ID, &user.FullName, &user.Email, &user.Role, &user.IsPINEnabled, &user.Currency, &user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email sudah terdaftar"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat akun pengguna"})
	}

	if err := SeedDefaultUserResources(ctx, tx, user.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyiapkan dompet & kategori default"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan akun pengguna"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	tag, err := h.DB.Exec(context.Background(),
		`DELETE FROM users WHERE id = $1 AND role = 'USER'`,
		id,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Akun pengguna tidak ditemukan"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
