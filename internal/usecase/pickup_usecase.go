package usecase

import (
	"context"
	"strings"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type PickupUsecase struct {
	pickupRepo *repository.PickupRepository
}

func NewPickupUsecase(
	pickupRepo *repository.PickupRepository,
) *PickupUsecase {
	return &PickupUsecase{pickupRepo: pickupRepo}
}

func (u *PickupUsecase) ListAvailable(ctx context.Context) ([]entity.PickupTask, error) {
	return u.pickupRepo.FindAvailable(ctx)
}

func (u *PickupUsecase) ListMine(ctx context.Context, volunteerID int) ([]entity.PickupTask, error) {
	if volunteerID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	return u.pickupRepo.FindByVolunteerID(ctx, volunteerID)
}

func (u *PickupUsecase) Claim(ctx context.Context, volunteerID, distributionID int) (*entity.Pickup, error) {
	if volunteerID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	if distributionID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID distribusi tidak valid")
	}

	return u.pickupRepo.Claim(ctx, volunteerID, distributionID)
}

func (u *PickupUsecase) UpdateStatus(
	ctx context.Context,
	volunteerID, pickupID int,
	status, notes string,
) (*entity.Pickup, error) {
	status = strings.TrimSpace(status)
	notes = strings.TrimSpace(notes)

	if volunteerID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID pengguna tidak valid")
	}

	if pickupID <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID tugas tidak valid")
	}

	switch status {
	case entity.PickupPickedUp, entity.PickupDelivered, entity.PickupCancelled:
	default:
		return nil, entity.NewDomainError(
			entity.ErrInvalid,
			"status harus picked_up, delivered, atau cancelled",
		)
	}

	return u.pickupRepo.UpdateStatus(ctx, volunteerID, pickupID, status, notes)
}