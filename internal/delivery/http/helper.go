package http

import (
	"errors"
	"log"

	"food-rescue/internal/entity"
	"food-rescue/pkg"

	"github.com/gofiber/fiber/v2"
)

// Helper balasan untuk semua route. Format JSON-nya diatur di pkg/response.go
// supaya seluruh endpoint memakai bentuk balasan yang sama.

func respondOK(c *fiber.Ctx, status int, message string, data any) error {
	return pkg.Success(c, status, message, data)
}

func respondMessage(c *fiber.Ctx, status int, message string) error {
	return pkg.Error(c, status, message)
}

// respondError memetakan error usecase ke status code HTTP.
func respondError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalid):
		return respondMessage(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrUnauthorized):
		return respondMessage(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, entity.ErrForbidden):
		return respondMessage(c, fiber.StatusForbidden, err.Error())
	case errors.Is(err, entity.ErrNotFound):
		return respondMessage(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrConflict):
		return respondMessage(c, fiber.StatusConflict, err.Error())
	default:
		log.Printf("error tidak terduga: %v", err)
		return respondMessage(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
}

// currentUserID membaca ID pengguna yang disimpan AuthMiddleware.
func currentUserID(c *fiber.Ctx) (int, bool) {
	userID, ok := c.Locals("userID").(int)

	return userID, ok
}