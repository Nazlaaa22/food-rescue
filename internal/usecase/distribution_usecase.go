package usecase

import (
	"context"
	"strings"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type DistributionUsecase struct {
	distributionRepo *repository.DistributionRepository
}

func NewDistributionUsecase(
	distributionRepo *repository.DistributionRepository,
) *DistributionUsecase {
	return &DistributionUsecase{distributionRepo: distributionRepo}
}

type AllocateInput struct {
	DonationID        int
	PartnerID         int
	PortionsAllocated int
}

// Allocate mengalokasikan sebagian porsi donasi ke satu mitra.
// Untuk membagi satu donasi ke beberapa mitra, Admin memanggil Allocate beberapa kali.
// Aturan sisa porsi, status donasi, dan keaktifan mitra diperiksa di dalam transaksi
// repository supaya aman dari alokasi bersamaan.
func (u *DistributionUsecase) Allocate(
	ctx context.Context,
	adminID int,
	input AllocateInput,
) (*entity.Distribution, error) {
	if adminID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID admin tidak valid")
	}

	if input.DonationID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "donasi wajib dipilih")
	}

	if input.PartnerID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "mitra penerima wajib dipilih")
	}

	if input.PortionsAllocated <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "jumlah porsi harus lebih dari 0")
	}

	distribution := &entity.Distribution{
		DonationID:        input.DonationID,
		PartnerID:         input.PartnerID,
		AdminID:           adminID,
		PortionsAllocated: input.PortionsAllocated,
	}

	if err := u.distributionRepo.Allocate(ctx, distribution); err != nil {
		return nil, err
	}

	return distribution, nil
}

func (u *DistributionUsecase) List(
	ctx context.Context,
	filter entity.DistributionFilter,
) ([]entity.DistributionDetail, error) {
	filter.Status = strings.TrimSpace(filter.Status)

	if filter.Status != "" && !entity.IsValidDistributionStatus(filter.Status) {
		return nil, entity.NewDomainError(entity.ErrInvalid, "status distribusi tidak valid")
	}

	if filter.DonationID < 0 || filter.PartnerID < 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "filter ID tidak valid")
	}

	return u.distributionRepo.FindAll(ctx, filter)
}

func (u *DistributionUsecase) GetByID(ctx context.Context, id int) (*entity.DistributionDetail, error) {
	if id <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID distribusi tidak valid")
	}

	return u.distributionRepo.FindByID(ctx, id)
}

// Cancel membatalkan distribusi yang belum diambil relawan sehingga porsinya
// kembali bisa dialokasikan ke mitra lain.
func (u *DistributionUsecase) Cancel(ctx context.Context, adminID, id int) (*entity.Distribution, error) {
	if adminID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID admin tidak valid")
	}

	if id <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID distribusi tidak valid")
	}

	return u.distributionRepo.Cancel(ctx, id, adminID)
}
