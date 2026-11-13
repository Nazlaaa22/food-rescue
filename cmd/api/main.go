package main

import (
	"log"

	"food-rescue/config"
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
	log.Println("Food Rescue API berhasil dijalankan.")
}
