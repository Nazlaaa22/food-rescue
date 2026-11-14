package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"food-rescue/internal/entity"
)

type PartnerRepository struct {
	db *sql.DB
}

func NewPartnerRepository(db *sql.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}

func (r *PartnerRepository) Create(ctx context.Context, partner *entity.Partner) error {
	query := `
		INSERT INTO partners (name, type, address, phone, is_active)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		partner.Name,
		partner.Type,
		partner.Address,
		partner.Phone,
		partner.IsActive,
	).Scan(&partner.ID)

	if err != nil {
		return fmt.Errorf("gagal membuat mitra: %w", err)
	}

	return nil
}

// FindAll mengambil daftar mitra. activeOnly = true hanya mengambil mitra aktif.
func (r *PartnerRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.Partner, error) {
	query := `
		SELECT id, name, type, address, COALESCE(phone, ''), COALESCE(is_active, FALSE)
		FROM partners
		WHERE ($1::boolean = FALSE OR is_active = TRUE)
		ORDER BY name ASC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar mitra: %w", err)
	}
	defer rows.Close()

	partners := []entity.Partner{}

	for rows.Next() {
		var partner entity.Partner

		err := rows.Scan(
			&partner.ID,
			&partner.Name,
			&partner.Type,
			&partner.Address,
			&partner.Phone,
			&partner.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca data mitra: %w", err)
		}

		partners = append(partners, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca data mitra: %w", err)
	}

	return partners, nil
}

func (r *PartnerRepository) FindByID(ctx context.Context, id int) (*entity.Partner, error) {
	query := `
		SELECT id, name, type, address, COALESCE(phone, ''), COALESCE(is_active, FALSE)
		FROM partners
		WHERE id = $1
	`

	partner := &entity.Partner{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&partner.ID,
		&partner.Name,
		&partner.Type,
		&partner.Address,
		&partner.Phone,
		&partner.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "mitra tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data mitra: %w", err)
	}

	return partner, nil
}

func (r *PartnerRepository) Update(ctx context.Context, partner *entity.Partner) error {
	query := `
		UPDATE partners
		SET name = $1, type = $2, address = $3, phone = NULLIF($4, ''), is_active = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		partner.Name,
		partner.Type,
		partner.Address,
		partner.Phone,
		partner.IsActive,
		partner.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal memperbarui mitra: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal memeriksa hasil pembaruan mitra: %w", err)
	}

	if affected == 0 {
		return entity.NewDomainError(entity.ErrNotFound, "mitra tidak ditemukan")
	}

	return nil
}
