package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

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

func validateNameAndPhone(name, phone string) error {
	if name == "" {
		return entity.NewDomainError(entity.ErrInvalid, "nama wajib diisi")
	}

	if utf8.RuneCountInString(name) > 100 {
		return entity.NewDomainError(entity.ErrInvalid, "nama maksimal 100 karakter")
	}

	if utf8.RuneCountInString(phone) > 20 {
		return entity.NewDomainError(entity.ErrInvalid, "nomor telepon maksimal 20 karakter")
	}

	return nil
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
		return nil, entity.NewDomainError(entity.ErrInvalid, "nama, email, dan password wajib diisi")
	}

	if err := validateNameAndPhone(name, phone); err != nil {
		return nil, err
	}

	if _, err := mail.ParseAddress(email); err != nil || len(email) > 255 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "format email tidak valid")
	}

	if len(password) < 8 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "password minimal 8 karakter")
	}

	// Admin tidak boleh mendaftarkan dirinya sendiri.
	if role != "donor" && role != "volunteer" {
		return nil, entity.NewDomainError(entity.ErrInvalid, "role harus donor atau volunteer")
	}

	// Memeriksa apakah email sudah digunakan.
	_, err := u.userRepo.FindByEmail(ctx, email)

	if err == nil {
		return nil, entity.NewDomainError(entity.ErrConflict, "email sudah terdaftar")
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
		return nil, err
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
		return nil, "", entity.NewDomainError(entity.ErrInvalid, "email dan password wajib diisi")
	}

	// Pesan untuk email tidak terdaftar dan password salah sengaja disamakan
	// supaya orang lain tidak bisa menebak email mana yang terdaftar.
	invalidCredentials := entity.NewDomainError(entity.ErrUnauthorized, "email atau password salah")

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", invalidCredentials
		}

		return nil, "", fmt.Errorf("gagal mencari pengguna: %w", err)
	}

	if !pkg.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", invalidCredentials
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

func (u *UserUsecase) GetProfile(ctx context.Context, userID int) (*entity.User, error) {
	if userID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "pengguna tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil profil: %w", err)
	}

	return user, nil
}

// UpdateProfile hanya mengubah nama dan nomor telepon.
// Email, role, dan status verifikasi tidak bisa diubah sendiri oleh pengguna.
func (u *UserUsecase) UpdateProfile(
	ctx context.Context,
	userID int,
	name, phone string,
) (*entity.User, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)

	if err := validateNameAndPhone(name, phone); err != nil {
		return nil, err
	}

	user, err := u.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = name
	user.Phone = phone

	err = u.userRepo.UpdateProfile(ctx, user)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "pengguna tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ListUsers dipakai admin untuk memantau pengguna, misalnya
// mencari akun yang belum diverifikasi (?is_verified=false).
func (u *UserUsecase) ListUsers(ctx context.Context, role string, verified *bool) ([]entity.User, error) {
	role = strings.ToLower(strings.TrimSpace(role))

	switch role {
	case "", "donor", "volunteer", "admin":
	default:
		return nil, entity.NewDomainError(entity.ErrInvalid, "role tidak valid")
	}

	return u.userRepo.FindAll(ctx, role, verified)
}

func (u *UserUsecase) VerifyUser(
	ctx context.Context,
	userID int,
	isVerified bool,
) error {
	if userID <= 0 {
		return entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	err := u.userRepo.UpdateVerificationStatus(ctx, userID, isVerified)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.NewDomainError(entity.ErrNotFound, "pengguna tidak ditemukan atau berstatus admin")
	}
	if err != nil {
		return fmt.Errorf("gagal memverifikasi pengguna: %w", err)
	}

	return nil
}
