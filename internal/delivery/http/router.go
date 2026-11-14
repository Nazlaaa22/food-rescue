package http

import (
	"database/sql"

	"food-rescue/config"
	"food-rescue/internal/middleware"
	"food-rescue/internal/repository"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, cfg *config.Config, db *sql.DB) {
	auth := middleware.AuthMiddleware(cfg.JWTSecret)
	verified := middleware.RequireVerified(db)

	// ===== Orang 1: user, auth, profile =====
	userUC := usecase.NewUserUsecase(repository.NewUserRepository(db), cfg.JWTSecret)

	RegisterUserRoutes(app, auth, userUC)

	// ===== Orang 2: food_categories, donations, pickups =====
	categoryUC := usecase.NewFoodCategoryUsecase(repository.NewFoodCategoryRepository(db))
	donationUC := usecase.NewDonationUsecase(repository.NewDonationRepository(db))
	pickupUC := usecase.NewPickupUsecase(repository.NewPickupRepository(db))

	RegisterFoodCategoryRoutes(app, auth, categoryUC)
	RegisterDonationRoutes(app, auth, verified, donationUC)
	RegisterPickupRoutes(app, auth, verified, pickupUC)

	// ===== Orang 3: partners, distributions, stats (tambahkan di sini) =====
}
