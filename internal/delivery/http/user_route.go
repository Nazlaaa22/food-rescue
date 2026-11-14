package http

import (
	"strconv"

	"food-rescue/internal/middleware"
	"food-rescue/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type userHandler struct {
	uc *usecase.UserUsecase
}

// RegisterUserRoutes mendaftarkan:
//
//	POST /auth/register            publik, body: {"name","email","password","phone","role": "donor"|"volunteer"}
//	POST /auth/login               publik, body: {"email","password"}
//	GET  /profile                  semua role yang login
//	PUT  /profile                  semua role yang login, body: {"name","phone"}
//	GET  /admin/users              admin (opsional ?role=donor&is_verified=false)
//	PUT  /admin/users/:id/verify   admin, body: {"is_verified": true|false} (default true)
func RegisterUserRoutes(
	app fiber.Router,
	auth fiber.Handler,
	uc *usecase.UserUsecase,
) {
	h := &userHandler{uc: uc}

	app.Post("/auth/register", h.register)
	app.Post("/auth/login", h.login)

	app.Get("/profile", auth, h.getProfile)
	app.Put("/profile", auth, h.updateProfile)

	admin := app.Group("/admin/users", auth, middleware.RequireRole("admin"))
	admin.Get("/", h.list)
	admin.Put("/:id/verify", h.verify)
}

func (h *userHandler) register(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	user, err := h.uc.Register(c.UserContext(), req.Name, req.Email, req.Password, req.Phone, req.Role)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusCreated,
		"Registrasi berhasil, tunggu verifikasi admin sebelum membuat donasi atau mengambil tugas", user)
}

func (h *userHandler) login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	user, token, err := h.uc.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *userHandler) getProfile(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	user, err := h.uc.GetProfile(c.UserContext(), userID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Profil pengguna", user)
}

func (h *userHandler) updateProfile(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return respondMessage(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan")
	}

	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}

	if err := c.BodyParser(&req); err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	user, err := h.uc.UpdateProfile(c.UserContext(), userID, req.Name, req.Phone)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Profil berhasil diperbarui", user)
}

func (h *userHandler) list(c *fiber.Ctx) error {
	var verified *bool

	if raw := c.Query("is_verified"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return respondMessage(c, fiber.StatusBadRequest, "is_verified harus true atau false")
		}

		verified = &value
	}

	users, err := h.uc.ListUsers(c.UserContext(), c.Query("role"), verified)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.StatusOK, "Daftar pengguna", users)
}

func (h *userHandler) verify(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return respondMessage(c, fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	var req struct {
		IsVerified *bool `json:"is_verified"`
	}

	// Body boleh kosong: tanpa is_verified berarti menyetujui akun.
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return respondMessage(c, fiber.StatusBadRequest, "Format request tidak valid")
		}
	}

	isVerified := true
	if req.IsVerified != nil {
		isVerified = *req.IsVerified
	}

	if err := h.uc.VerifyUser(c.UserContext(), userID, isVerified); err != nil {
		return respondError(c, err)
	}

	message := "Akun berhasil diverifikasi"
	if !isVerified {
		message = "Verifikasi akun dibatalkan"
	}

	return respondOK(c, fiber.StatusOK, message, fiber.Map{
		"user_id":     userID,
		"is_verified": isVerified,
	})
}
