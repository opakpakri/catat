package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthTenant(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization header",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok || int(userIDFloat) <= 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user identity in token",
			})
		}

		c.Locals("userID", int(userIDFloat))
		return c.Next()
	}
}

func AdminOnly(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		var role string
		err := db.QueryRow(context.Background(), `SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
		if err != nil || role != "ADMIN" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Akses ditolak. Hanya Admin yang dapat mengakses fitur ini.",
			})
		}
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) int {
	id, _ := c.Locals("userID").(int)
	return id
}
