package usecase

import (
	"context"
	"strings"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type DonationLogUsecase struct {
	donationLogRepo *repository.DonationLogRepository
}

func NewDonationLogUsecase(
	donationLogRepo *repository.DonationLogRepository,
) *DonationLogUsecase {
	return &DonationLogUsecase{
		donationLogRepo: donationLogRepo,
	}
}

func (u *DonationLogUsecase) CreateLog(
	ctx context.Context,
	donationID int,
	status string,
	changedBy int,
) error {
	status = strings.TrimSpace(status)

	if donationID <= 0 {
		return entity.NewDomainError(entity.ErrInvalid, "ID donasi tidak valid")
	}

	if !entity.IsValidDonationStatus(status) {
		return entity.NewDomainError(entity.ErrInvalid, "status donasi tidak valid")
	}

	if changedBy <= 0 {
		return entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	log := &entity.DonationLog{
		DonationID: donationID,
		Status:     status,
		ChangedBy:  changedBy,
	}

	return u.donationLogRepo.Create(ctx, log)
}

func (u *DonationLogUsecase) GetLogsByDonationID(
	ctx context.Context,
	donationID int,
) ([]entity.DonationLog, error) {
	if donationID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID donasi tidak valid")
	}

	return u.donationLogRepo.FindByDonationID(ctx, donationID)
}
