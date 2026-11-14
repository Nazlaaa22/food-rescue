// Alat bantu SEMENTARA untuk testing lokal, jangan di-commit.
// Dipakai karena endpoint login belum ada.
//
// Contoh: go run ./cmd/gentoken 2 donor
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"food-rescue/config"
	"food-rescue/pkg"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("pemakaian: go run ./cmd/gentoken <user_id> <donor|volunteer|admin>")
	}

	userID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("user_id harus angka: %v", err)
	}

	cfg := config.LoadConfig()

	token, err := pkg.GenerateJWT(userID, os.Args[2], cfg.JWTSecret)
	if err != nil {
		log.Fatalf("gagal membuat token: %v", err)
	}

	fmt.Println(token)
}