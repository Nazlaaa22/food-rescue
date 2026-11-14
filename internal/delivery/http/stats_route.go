package http

import (
	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type statsHandler struct {
	uc *usecase.StatsUsecase
}

// RegisterStatsRoutes mendaftarkan:
//
//	GET /admin/stats   ringkasan donasi, distribusi, mitra, dan pengguna (admin)
func RegisterStatsRoutes(
	app fiber.Router,
	auth fiber.Handler,
	uc *usecase.StatsUsecase,
) {
	h := &statsHandler{uc: uc}

	app.Get("/admin/stats", auth, middleware.RequireRole("admin"), h.get)
}

func (h *statsHandler) get(c *fiber.Ctx) error {
	stats, err := h.uc.Get(c.UserContext())
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Statistik sistem", stats)
}
