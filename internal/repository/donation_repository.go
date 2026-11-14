package repository

import (
	"context"
	"database/sql"
	"fmt"

	"food-rescue/internal/entity"
)

type DonationRepository struct {
	db *sql.DB
}

func NewDonationRepository(db *sql.DB) *DonationRepository {
	return &DonationRepository{db: db}
}

// rowScanner menyatukan *sql.Row dan *sql.Rows supaya fungsi scan bisa dipakai bersama.
type rowScanner interface {
	Scan(dest ...any) error
}

const donationSelect = `
	SELECT id, donor_id, category_id, title, COALESCE(description, ''), portions,
	       expired_at, pickup_address, status, created_at
	FROM donations
`

// Create menyimpan donasi baru (status available) sekaligus mencatat
// riwayat pertamanya di donation_logs dalam satu transaksi.
func (r *DonationRepository) Create(ctx context.Context, donation *entity.Donation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO donations (
			donor_id, category_id, title, description, portions, expired_at, pickup_address
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
		RETURNING id, status, created_at
	`,
		donation.DonorID,
		donation.CategoryID,
		donation.Title,
		donation.Description,
		donation.Portions,
		donation.ExpiredAt,
		donation.PickupAddress,
	).Scan(&donation.ID, &donation.Status, &donation.CreatedAt)

	if err != nil {
		// 23503 = foreign key violation, di sini berarti category_id tidak ada.
		if hasPgErrorCode(err, "23503") {
			return entity.NewDomainError(entity.ErrInvalid, "kategori makanan tidak ditemukan")
		}

		return fmt.Errorf("gagal membuat donasi: %w", err)
	}

	if err := insertDonationLog(ctx, tx, donation.ID, donation.Status, donation.DonorID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal menyimpan donasi: %w", err)
	}

	return nil
}

func (r *DonationRepository) FindByDonorID(ctx context.Context, donorID int) ([]entity.Donation, error) {
	query := donationSelect + `
		WHERE donor_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, donorID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat donasi: %w", err)
	}
	defer rows.Close()

	return scanDonations(rows)
}

// FindAll mengambil semua donasi. Status kosong berarti tanpa filter.
func (r *DonationRepository) FindAll(ctx context.Context, status string) ([]entity.Donation, error) {
	query := donationSelect + `
		WHERE ($1::text = '' OR status = $1::text)
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar donasi: %w", err)
	}
	defer rows.Close()

	return scanDonations(rows)
}

func scanDonations(rows *sql.Rows) ([]entity.Donation, error) {
	donations := []entity.Donation{}

	for rows.Next() {
		var d entity.Donation

		err := rows.Scan(
			&d.ID,
			&d.DonorID,
			&d.CategoryID,
			&d.Title,
			&d.Description,
			&d.Portions,
			&d.ExpiredAt,
			&d.PickupAddress,
			&d.Status,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca donasi: %w", err)
		}

		donations = append(donations, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca donasi: %w", err)
	}

	return donations, nil
}

// insertDonationLog mencatat perubahan status donasi di dalam transaksi yang sedang berjalan.
// Dipakai juga oleh pickup_repository.go dan distribution_repository.go.
func insertDonationLog(ctx context.Context, tx *sql.Tx, donationID int, status string, changedBy int) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO donation_logs (donation_id, status, changed_by)
		VALUES ($1, $2, $3)
	`, donationID, status, changedBy)
	if err != nil {
		return fmt.Errorf("gagal mencatat riwayat donasi: %w", err)
	}

	return nil
}
