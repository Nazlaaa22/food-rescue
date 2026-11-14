package middleware

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

// RequireRole hanya mengizinkan request dari role yang terdaftar.
// Harus dipasang setelah AuthMiddleware karena membaca role dari c.Locals.
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token autentikasi tidak ditemukan",
			})
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Anda tidak memiliki akses ke resource ini",
		})
	}
}

// RequireVerified hanya mengizinkan akun yang sudah diverifikasi admin.
// JWT tidak membawa is_verified, jadi statusnya dicek ke database setiap request.
// Harus dipasang setelah AuthMiddleware.
func RequireVerified(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token autentikasi tidak ditemukan",
			})
		}

		var isVerified bool

		err := db.QueryRowContext(
			c.UserContext(),
			`SELECT COALESCE(is_verified, FALSE) FROM users WHERE id = $1`,
			userID,
		).Scan(&isVerified)

		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Pengguna tidak ditemukan",
			})
		}

		if err != nil {
			log.Printf("gagal memeriksa verifikasi akun: %v", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Terjadi kesalahan pada server",
			})
		}

		if !isVerified {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Akun Anda belum diverifikasi oleh admin",
			})
		}

		return c.Next()
	}
}