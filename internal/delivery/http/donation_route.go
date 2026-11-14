package http

import (
	"time"

	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type donationHandler struct {
	uc *usecase.DonationUsecase
}

// RegisterDonationRoutes mendaftarkan:
//
//	POST /donations         donatur (akun harus terverifikasi)
//	GET  /donations         donatur, riwayat donasinya sendiri
//	GET  /admin/donations   admin, semua donasi (opsional ?status=available)
func RegisterDonationRoutes(
	app fiber.Router,
	auth fiber.Handler,
	verified fiber.Handler,
	uc *usecase.DonationUsecase,
) {
	h := &donationHandler{uc: uc}

	donor := app.Group("/donations", auth, middleware.RequireRole("donor"))
	donor.Post("/", verified, h.create)
	donor.Get("/", h.listMine)

	app.Get("/admin/donations", auth, middleware.RequireRole("admin"), h.listAll)
}

func (h *donationHandler) create(c *fiber.Ctx) error {
	donorID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	var req struct {
		CategoryID    int       `json:"category_id"`
		Title         string    `json:"title"`
		Description   string    `json:"description"`
		Portions      int       `json:"portions"`
		ExpiredAt     time.Time `json:"expired_at"`
		PickupAddress string    `json:"pickup_address"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest,
			"Format request tidak valid (expired_at memakai format RFC3339, contoh 2026-12-01T18:00:00+07:00)")
	}

	donation, err := h.uc.Create(c.UserContext(), donorID, usecase.CreateDonationInput{
		CategoryID:    req.CategoryID,
		Title:         req.Title,
		Description:   req.Description,
		Portions:      req.Portions,
		ExpiredAt:     req.ExpiredAt,
		PickupAddress: req.PickupAddress,
	})
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Donasi berhasil dibuat", donation)
}

func (h *donationHandler) listMine(c *fiber.Ctx) error {
	donorID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	donations, err := h.uc.ListByDonor(c.UserContext(), donorID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Riwayat donasi", donations)
}

func (h *donationHandler) listAll(c *fiber.Ctx) error {
	donations, err := h.uc.ListAll(c.UserContext(), c.Query("status"))
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar donasi", donations)
}