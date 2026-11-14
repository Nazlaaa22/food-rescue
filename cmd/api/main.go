package main

import (
	"log"

	"food-rescue/config"
	delivery "food-rescue/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Membaca konfigurasi aplikasi
	cfg := config.LoadConfig()

	// Menghubungkan aplikasi ke PostgreSQL
	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("Gagal menghubungkan database: %v", err)
	}
	defer db.Close()

	log.Println("Berhasil terhubung ke PostgreSQL!")

	app := fiber.New()
	delivery.SetupRouter(app, cfg, db)

	log.Printf("Food Rescue API berjalan di port %s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}