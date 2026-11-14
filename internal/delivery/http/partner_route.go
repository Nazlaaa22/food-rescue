package http

import (
	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type partnerHandler struct {
	uc *usecase.PartnerUsecase
}

type partnerRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}

func (r partnerRequest) toInput() usecase.PartnerInput {
	return usecase.PartnerInput{
		Name:     r.Name,
		Type:     r.Type,
		Address:  r.Address,
		Phone:    r.Phone,
		IsActive: r.IsActive,
	}
}

// RegisterPartnerRoutes mendaftarkan (semua khusus admin):
//
//	GET  /admin/partners       daftar mitra (opsional ?active=true)
//	GET  /admin/partners/:id   detail mitra
//	POST /admin/partners       body: {"name","type","address","phone","is_active"}
//	PUT  /admin/partners/:id   ubah data atau nonaktifkan mitra
func RegisterPartnerRoutes(
	app fiber.Router,
	auth fiber.Handler,
	uc *usecase.PartnerUsecase,
) {
	h := &partnerHandler{uc: uc}

	admin := app.Group("/admin/partners", auth, middleware.RequireRole("admin"))
	admin.Get("/", h.list)
	admin.Get("/:id", h.detail)
	admin.Post("/", h.create)
	admin.Put("/:id", h.update)
}

func (h *partnerHandler) list(c *fiber.Ctx) error {
	activeOnly := c.Query("active") == "true"

	partners, err := h.uc.List(c.UserContext(), activeOnly)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar mitra penerima", partners)
}

func (h *partnerHandler) detail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID mitra tidak valid")
	}

	partner, err := h.uc.GetByID(c.UserContext(), id)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Detail mitra penerima", partner)
}

func (h *partnerHandler) create(c *fiber.Ctx) error {
	var req partnerRequest

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	partner, err := h.uc.Create(c.UserContext(), req.toInput())
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Mitra berhasil ditambahkan", partner)
}

func (h *partnerHandler) update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID mitra tidak valid")
	}

	var req partnerRequest

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	partner, err := h.uc.Update(c.UserContext(), id, req.toInput())
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Data mitra berhasil diperbarui", partner)
}
