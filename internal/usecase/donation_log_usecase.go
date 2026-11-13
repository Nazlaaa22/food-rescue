package usecase

import (
	"context"
	"errors"
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
		return errors.New("ID donasi tidak valid")
	}

	if status == "" {
		return errors.New("status donasi tidak boleh kosong")
	}

	if changedBy <= 0 {
		return errors.New("ID pengguna tidak valid")
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
		return nil, errors.New("ID donasi tidak valid")
	}

	return u.donationLogRepo.FindByDonationID(ctx, donationID)
}
