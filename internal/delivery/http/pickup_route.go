package http

import (
	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type pickupHandler struct {
	uc *usecase.PickupUsecase
}

// RegisterPickupRoutes mendaftarkan (semua untuk relawan terverifikasi):
//
//	GET /pickups/available   tugas antar yang belum diambil
//	GET /pickups/mine        riwayat tugas relawan
//	POST /pickups            ambil tugas, body: {"distribution_id": 1}
//	PUT /pickups/:id/status  body: {"status": "picked_up" | "delivered" | "cancelled", "notes": "..."}
func RegisterPickupRoutes(
	app fiber.Router,
	auth fiber.Handler,
	verified fiber.Handler,
	uc *usecase.PickupUsecase,
) {
	h := &pickupHandler{uc: uc}

	volunteer := app.Group("/pickups", auth, middleware.RequireRole("volunteer"), verified)
	volunteer.Get("/available", h.listAvailable)
	volunteer.Get("/mine", h.listMine)
	volunteer.Post("/", h.claim)
	volunteer.Put("/:id/status", h.updateStatus)
}

func (h *pickupHandler) listAvailable(c *fiber.Ctx) error {
	tasks, err := h.uc.ListAvailable(c.UserContext())
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar tugas antar yang tersedia", tasks)
}

func (h *pickupHandler) listMine(c *fiber.Ctx) error {
	volunteerID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	tasks, err := h.uc.ListMine(c.UserContext(), volunteerID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Riwayat tugas", tasks)
}

func (h *pickupHandler) claim(c *fiber.Ctx) error {
	volunteerID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	var req struct {
		DistributionID int `json:"distribution_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	pickup, err := h.uc.Claim(c.UserContext(), volunteerID, req.DistributionID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Tugas berhasil diambil", pickup)
}

func (h *pickupHandler) updateStatus(c *fiber.Ctx) error {
	volunteerID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	pickupID, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID tugas tidak valid")
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	pickup, err := h.uc.UpdateStatus(c.UserContext(), volunteerID, pickupID, req.Status, req.Notes)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Status tugas berhasil diperbarui", pickup)
}