package usecase

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type DonationUsecase struct {
	donationRepo *repository.DonationRepository
}

func NewDonationUsecase(
	donationRepo *repository.DonationRepository,
) *DonationUsecase {
	return &DonationUsecase{donationRepo: donationRepo}
}

type CreateDonationInput struct {
	CategoryID    int
	Title         string
	Description   string
	Portions      int
	ExpiredAt     time.Time
	PickupAddress string
}

func (u *DonationUsecase) Create(
	ctx context.Context,
	donorID int,
	input CreateDonationInput,
) (*entity.Donation, error) {
	title := strings.TrimSpace(input.Title)
	pickupAddress := strings.TrimSpace(input.PickupAddress)

	if donorID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	if input.CategoryID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "kategori makanan wajib dipilih")
	}

	if title == "" {
		return nil, entity.NewDomainError(entity.ErrInvalid, "nama makanan wajib diisi")
	}

	if utf8.RuneCountInString(title) > 150 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "nama makanan maksimal 150 karakter")
	}

	if input.Portions <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "jumlah porsi harus lebih dari 0")
	}

	if pickupAddress == "" {
		return nil, entity.NewDomainError(entity.ErrInvalid, "lokasi pengambilan wajib diisi")
	}

	if input.ExpiredAt.IsZero() {
		return nil, entity.NewDomainError(entity.ErrInvalid, "batas waktu layak konsumsi wajib diisi")
	}

	// Kolom expired_at bertipe TIMESTAMP tanpa zona waktu,
	// jadi selalu disimpan dalam UTC supaya perbandingannya konsisten.
	expiredAt := input.ExpiredAt.UTC()

	if !expiredAt.After(time.Now().UTC()) {
		return nil, entity.NewDomainError(entity.ErrInvalid, "batas waktu layak konsumsi harus di masa depan")
	}

	donation := &entity.Donation{
		DonorID:       donorID,
		CategoryID:    input.CategoryID,
		Title:         title,
		Description:   strings.TrimSpace(input.Description),
		Portions:      input.Portions,
		ExpiredAt:     expiredAt,
		PickupAddress: pickupAddress,
	}

	if err := u.donationRepo.Create(ctx, donation); err != nil {
		return nil, err
	}

	return donation, nil
}

func (u *DonationUsecase) ListByDonor(ctx context.Context, donorID int) ([]entity.Donation, error) {
	if donorID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	return u.donationRepo.FindByDonorID(ctx, donorID)
}

// ListAll dipakai admin. Status kosong berarti semua donasi.
func (u *DonationUsecase) ListAll(ctx context.Context, status string) ([]entity.Donation, error) {
	status = strings.TrimSpace(status)

	if status != "" && !entity.IsValidDonationStatus(status) {
		return nil, entity.NewDomainError(entity.ErrInvalid, "status donasi tidak valid")
	}

	return u.donationRepo.FindAll(ctx, status)
}