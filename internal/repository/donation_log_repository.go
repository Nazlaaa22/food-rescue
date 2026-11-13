package repository

import (
	"context"
	"database/sql"
	"fmt"

	"food-rescue/internal/entity"
)

type DonationLogRepository struct {
	db *sql.DB
}

func NewDonationLogRepository(db *sql.DB) *DonationLogRepository {
	return &DonationLogRepository{db: db}
}

func (r *DonationLogRepository) Create(ctx context.Context, log *entity.DonationLog) error {
	query := `
		INSERT INTO donation_logs (donation_id, status, changed_by)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		log.DonationID,
		log.Status,
		log.ChangedBy,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("gagal menyimpan riwayat donasi: %w", err)
	}

	return nil
}

func (r *DonationLogRepository) FindByDonationID(ctx context.Context, donationID int) ([]entity.DonationLog, error) {
	query := `
		SELECT id, donation_id, status, changed_by, created_at
		FROM donation_logs
		WHERE donation_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, donationID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat donasi: %w", err)
	}
	defer rows.Close()

	var logs []entity.DonationLog

	for rows.Next() {
		var log entity.DonationLog

		err := rows.Scan(
			&log.ID,
			&log.DonationID,
			&log.Status,
			&log.ChangedBy,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca riwayat donasi: %w", err)
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca riwayat donasi: %w", err)
	}

	return logs, nil
}
