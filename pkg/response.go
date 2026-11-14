package pkg

import "github.com/gofiber/fiber/v2"

// Response adalah format balasan standar seluruh endpoint Food Rescue.
//
//	{ "success": true,  "message": "Daftar mitra", "data": [...] }
//	{ "success": false, "message": "mitra tidak ditemukan" }
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success mengirim balasan berhasil beserta data.
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error mengirim balasan gagal tanpa data.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
	})
}
