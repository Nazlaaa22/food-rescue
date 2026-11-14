package http

import (
	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type foodCategoryHandler struct {
	uc *usecase.FoodCategoryUsecase
}

// RegisterFoodCategoryRoutes mendaftarkan:
//
//	GET  /categories         semua role yang login (donatur butuh daftar kategori)
//	GET  /admin/categories   admin
//	POST /admin/categories   admin
func RegisterFoodCategoryRoutes(
	app fiber.Router,
	auth fiber.Handler,
	uc *usecase.FoodCategoryUsecase,
) {
	h := &foodCategoryHandler{uc: uc}

	app.Get("/categories", auth, h.list)

	admin := app.Group("/admin/categories", auth, middleware.RequireRole("admin"))
	admin.Get("/", h.list)
	admin.Post("/", h.create)
}

func (h *foodCategoryHandler) list(c *fiber.Ctx) error {
	categories, err := h.uc.List(c.UserContext())
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar kategori makanan", categories)
}

func (h *foodCategoryHandler) create(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	category, err := h.uc.Create(c.UserContext(), req.Name)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated, "Kategori makanan berhasil dibuat", category)
}