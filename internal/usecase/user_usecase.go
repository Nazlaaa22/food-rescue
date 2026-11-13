package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
	"food-rescue/pkg"
)

type UserUsecase struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewUserUsecase(
	userRepo *repository.UserRepository,
	jwtSecret string,
) *UserUsecase {
	return &UserUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register digunakan untuk mendaftarkan pengguna baru.
func (u *UserUsecase) Register(
	ctx context.Context,
	name, email, password, phone, role string,
) (*entity.User, error) {

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)
	role = strings.ToLower(strings.TrimSpace(role))

	if name == "" || email == "" || password == "" {
		return nil, errors.New("nama, email, dan password wajib diisi")
	}

	if len(password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}

	// Admin tidak boleh mendaftarkan dirinya sendiri.
	if role != "donor" && role != "volunteer" {
		return nil, errors.New("role harus donor atau volunteer")
	}

	// Memeriksa apakah email sudah digunakan.
	_, err := u.userRepo.FindByEmail(ctx, email)

	if err == nil {
		return nil, errors.New("email sudah terdaftar")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("gagal memeriksa email: %w", err)
	}

	// Mengamankan password dengan bcrypt.
	hashedPassword, err := pkg.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat hash password: %w", err)
	}

	user := &entity.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Phone:        phone,
		Role:         role,
		IsVerified:   false,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan pengguna: %w", err)
	}

	return user, nil
}

// Login digunakan untuk memverifikasi pengguna dan membuat JWT.
func (u *UserUsecase) Login(
	ctx context.Context,
	email, password string,
) (*entity.User, string, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return nil, "", errors.New("email dan password wajib diisi")
	}

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("email atau password salah")
		}

		return nil, "", fmt.Errorf("gagal mencari pengguna: %w", err)
	}

	if !pkg.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", errors.New("email atau password salah")
	}

	token, err := pkg.GenerateJWT(
		user.ID,
		user.Role,
		u.jwtSecret,
	)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat token: %w", err)
	}

	return user, token, nil
}

func (u *UserUsecase) VerifyUser(
	ctx context.Context,
	userID int,
	isVerified bool,
) error {
	if userID <= 0 {
		return errors.New("ID pengguna tidak valid")
	}

	err := u.userRepo.UpdateVerificationStatus(ctx, userID, isVerified)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("pengguna tidak ditemukan")
	}
	if err != nil {
		return fmt.Errorf("gagal memverifikasi pengguna: %w", err)
	}

	return nil
}
