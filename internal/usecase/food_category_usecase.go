package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type FoodCategoryUsecase struct {
	categoryRepo *repository.FoodCategoryRepository
}

func NewFoodCategoryUsecase(
	categoryRepo *repository.FoodCategoryRepository,
) *FoodCategoryUsecase {
	return &FoodCategoryUsecase{categoryRepo: categoryRepo}
}

func (u *FoodCategoryUsecase) Create(ctx context.Context, name string) (*entity.FoodCategory, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, entity.NewDomainError(entity.ErrInvalid, "nama kategori wajib diisi")
	}

	if utf8.RuneCountInString(name) > 100 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "nama kategori maksimal 100 karakter")
	}

	category := &entity.FoodCategory{Name: name}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (u *FoodCategoryUsecase) List(ctx context.Context) ([]entity.FoodCategory, error) {
	return u.categoryRepo.FindAll(ctx)
}