package http

import (
	"food-rescue/internal/entity"
	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type distributionHandler struct {
	uc *usecase.DistributionUsecase
}

// RegisterDistributionRoutes mendaftarkan (semua khusus admin):
//
//	POST /admin/distributions             alokasi, body: {"donation_id","partner_id","portions_allocated"}
//	GET  /admin/distributions             pantau penyaluran (opsional ?status=&donation_id=&partner_id=)
//	GET  /admin/distributions/:id         detail distribusi beserta status pickup
//	PUT  /admin/distributions/:id/cancel  batalkan distribusi yang belum diambil relawan
func RegisterDistributionRoutes(
	app fiber.Router,
	auth fiber.Handler,
	uc *usecase.DistributionUsecase,
) {
	h := &distributionHandler{uc: uc}

	admin := app.Group("/admin/distributions", auth, middleware.RequireRole("admin"))
	admin.Post("/", h.allocate)
	admin.Get("/", h.list)
	admin.Get("/:id", h.detail)
	admin.Put("/:id/cancel", h.cancel)
}

func (h *distributionHandler) allocate(c *fiber.Ctx) error {
	adminID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	var req struct {
		DonationID        int `json:"donation_id"`
		PartnerID         int `json:"partner_id"`
		PortionsAllocated int `json:"portions_allocated"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	distribution, err := h.uc.Allocate(c.UserContext(), adminID, usecase.AllocateInput{
		DonationID:        req.DonationID,
		PartnerID:         req.PartnerID,
		PortionsAllocated: req.PortionsAllocated,
	})
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Donasi berhasil dialokasikan ke mitra", distribution)
}

func (h *distributionHandler) list(c *fiber.Ctx) error {
	filter := entity.DistributionFilter{
		Status:     c.Query("status"),
		DonationID: c.QueryInt("donation_id"),
		PartnerID:  c.QueryInt("partner_id"),
	}

	distributions, err := h.uc.List(c.UserContext(), filter)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar distribusi", distributions)
}

func (h *distributionHandler) detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID distribusi tidak valid")
	}

	distribution, err := h.uc.GetByID(c.UserContext(), id)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail distribusi", distribution)
}

func (h *distributionHandler) cancel(c *fiber.Ctx) error {
	adminID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID distribusi tidak valid")
	}

	distribution, err := h.uc.Cancel(c.UserContext(), adminID, id)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Distribusi berhasil dibatalkan", distribution)
}
