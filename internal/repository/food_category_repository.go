package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"food-rescue/internal/entity"

	"github.com/jackc/pgx/v5/pgconn"
)

type FoodCategoryRepository struct {
	db *sql.DB
}

func NewFoodCategoryRepository(db *sql.DB) *FoodCategoryRepository {
	return &FoodCategoryRepository{db: db}
}

// hasPgErrorCode memeriksa kode error PostgreSQL (23505 = unique, 23503 = foreign key).
// Dipakai juga oleh donation_repository.go.
func hasPgErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == code
}

func (r *FoodCategoryRepository) Create(ctx context.Context, category *entity.FoodCategory) error {
	query := `
		INSERT INTO food_categories (name)
		VALUES ($1)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, category.Name).Scan(&category.ID)
	if err != nil {
		if hasPgErrorCode(err, "23505") {
			return entity.NewDomainError(entity.ErrConflict, "nama kategori sudah digunakan")
		}

		return fmt.Errorf("gagal membuat kategori makanan: %w", err)
	}

	return nil
}

func (r *FoodCategoryRepository) FindAll(ctx context.Context) ([]entity.FoodCategory, error) {
	query := `
		SELECT id, name
		FROM food_categories
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil kategori makanan: %w", err)
	}
	defer rows.Close()

	categories := []entity.FoodCategory{}

	for rows.Next() {
		var category entity.FoodCategory

		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("gagal membaca kategori makanan: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca kategori makanan: %w", err)
	}

	return categories, nil
}