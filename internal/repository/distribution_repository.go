package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"food-rescue/internal/entity"
)

type DistributionRepository struct {
	db *sql.DB
}

func NewDistributionRepository(db *sql.DB) *DistributionRepository {
	return &DistributionRepository{db: db}
}

// Allocate membuat distribusi baru dalam satu transaksi:
//  1. mengunci baris donasi dan memastikan donasi masih bisa dialokasikan,
//  2. memastikan mitra ada dan aktif,
//  3. memastikan porsi yang diminta tidak melebihi sisa porsi donasi,
//  4. menyimpan distribusi, lalu mengubah donasi available -> allocated
//     dan mencatatnya ke donation_logs.
func (r *DistributionRepository) Allocate(ctx context.Context, d *entity.Distribution) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	var donationStatus string
	var totalPortions int
	var notExpired bool

	// Baris donasi dikunci supaya dua admin tidak mengalokasikan porsi yang sama bersamaan.
	err = tx.QueryRowContext(ctx, `
		SELECT status, portions, expired_at > (NOW() AT TIME ZONE 'UTC')
		FROM donations
		WHERE id = $1
		FOR UPDATE
	`, d.DonationID).Scan(&donationStatus, &totalPortions, &notExpired)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.NewDomainError(entity.ErrNotFound, "donasi tidak ditemukan")
	}
	if err != nil {
		return fmt.Errorf("gagal memeriksa donasi: %w", err)
	}

	if donationStatus != entity.DonationAvailable && donationStatus != entity.DonationAllocated {
		return entity.NewDomainError(
			entity.ErrConflict,
			fmt.Sprintf("donasi berstatus %s tidak dapat dialokasikan", donationStatus),
		)
	}

	if !notExpired {
		return entity.NewDomainError(entity.ErrConflict, "donasi sudah melewati batas waktu layak konsumsi")
	}

	var partnerActive bool

	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(is_active, FALSE) FROM partners WHERE id = $1`,
		d.PartnerID,
	).Scan(&partnerActive)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.NewDomainError(entity.ErrNotFound, "mitra tidak ditemukan")
	}
	if err != nil {
		return fmt.Errorf("gagal memeriksa mitra: %w", err)
	}

	if !partnerActive {
		return entity.NewDomainError(entity.ErrConflict, "mitra sedang tidak aktif")
	}

	var allocated int

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(portions_allocated), 0)
		FROM distributions
		WHERE donation_id = $1 AND status <> 'cancelled'
	`, d.DonationID).Scan(&allocated)
	if err != nil {
		return fmt.Errorf("gagal menghitung porsi teralokasi: %w", err)
	}

	remaining := totalPortions - allocated
	if d.PortionsAllocated > remaining {
		return entity.NewDomainError(
			entity.ErrConflict,
			fmt.Sprintf("porsi melebihi sisa donasi (sisa %d porsi)", remaining),
		)
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO distributions (donation_id, partner_id, admin_id, portions_allocated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, allocated_at
	`, d.DonationID, d.PartnerID, d.AdminID, d.PortionsAllocated).Scan(&d.ID, &d.Status, &d.AllocatedAt)
	if err != nil {
		return fmt.Errorf("gagal menyimpan distribusi: %w", err)
	}

	if donationStatus == entity.DonationAvailable {
		if err := setDonationStatus(ctx, tx, d.DonationID, entity.DonationAllocated, d.AdminID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal menyimpan alokasi: %w", err)
	}

	return nil
}

const distributionDetailSelect = `
	SELECT d.id, d.donation_id, d.partner_id, d.admin_id, d.portions_allocated,
	       d.status, d.allocated_at, dn.title, pt.name, p.volunteer_id, p.status
	FROM distributions d
	JOIN donations dn ON dn.id = d.donation_id
	JOIN partners pt ON pt.id = d.partner_id
	LEFT JOIN pickups p ON p.distribution_id = d.id AND p.status <> 'cancelled'
`

func (r *DistributionRepository) FindAll(
	ctx context.Context,
	filter entity.DistributionFilter,
) ([]entity.DistributionDetail, error) {
	var conditions []string
	var args []any

	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", len(args)))
	}

	if filter.DonationID > 0 {
		args = append(args, filter.DonationID)
		conditions = append(conditions, fmt.Sprintf("d.donation_id = $%d", len(args)))
	}

	if filter.PartnerID > 0 {
		args = append(args, filter.PartnerID)
		conditions = append(conditions, fmt.Sprintf("d.partner_id = $%d", len(args)))
	}

	query := distributionDetailSelect

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY d.allocated_at DESC, d.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar distribusi: %w", err)
	}
	defer rows.Close()

	details := []entity.DistributionDetail{}

	for rows.Next() {
		detail, err := scanDistributionDetail(rows)
		if err != nil {
			return nil, err
		}

		details = append(details, *detail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca distribusi: %w", err)
	}

	return details, nil
}

func (r *DistributionRepository) FindByID(ctx context.Context, id int) (*entity.DistributionDetail, error) {
	row := r.db.QueryRowContext(ctx, distributionDetailSelect+" WHERE d.id = $1", id)

	detail, err := scanDistributionDetail(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "distribusi tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	return detail, nil
}

func scanDistributionDetail(row rowScanner) (*entity.DistributionDetail, error) {
	var detail entity.DistributionDetail
	var volunteerID sql.NullInt64
	var pickupStatus sql.NullString

	err := row.Scan(
		&detail.ID,
		&detail.DonationID,
		&detail.PartnerID,
		&detail.AdminID,
		&detail.PortionsAllocated,
		&detail.Status,
		&detail.AllocatedAt,
		&detail.DonationTitle,
		&detail.PartnerName,
		&volunteerID,
		&pickupStatus,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("gagal membaca distribusi: %w", err)
	}

	if volunteerID.Valid {
		id := int(volunteerID.Int64)
		detail.VolunteerID = &id
	}

	if pickupStatus.Valid {
		detail.PickupStatus = &pickupStatus.String
	}

	return &detail, nil
}

// Cancel membatalkan distribusi yang belum diambil relawan, lalu menyelaraskan status donasi:
//   - tidak ada distribusi aktif dan belum ada yang terkirim -> donasi kembali available,
//   - tidak ada distribusi aktif tapi ada yang sudah delivered -> donasi menjadi delivered.
func (r *DistributionRepository) Cancel(ctx context.Context, id, adminID int) (*entity.Distribution, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	d := &entity.Distribution{ID: id}

	err = tx.QueryRowContext(ctx, `
		SELECT donation_id, partner_id, admin_id, portions_allocated, status, allocated_at
		FROM distributions
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&d.DonationID, &d.PartnerID, &d.AdminID, &d.PortionsAllocated, &d.Status, &d.AllocatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "distribusi tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil distribusi: %w", err)
	}

	if d.Status != entity.DistributionAllocated {
		return nil, entity.NewDomainError(
			entity.ErrConflict,
			fmt.Sprintf("distribusi berstatus %s tidak dapat dibatalkan", d.Status),
		)
	}

	var claimed bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pickups
			WHERE distribution_id = $1 AND status <> 'cancelled'
		)
	`, id).Scan(&claimed)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa pickup: %w", err)
	}

	if claimed {
		return nil, entity.NewDomainError(entity.ErrConflict, "distribusi sudah diambil relawan")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE distributions SET status = 'cancelled' WHERE id = $1`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membatalkan distribusi: %w", err)
	}

	d.Status = entity.DistributionCancelled

	var donationStatus string

	err = tx.QueryRowContext(ctx,
		`SELECT status FROM donations WHERE id = $1 FOR UPDATE`, d.DonationID,
	).Scan(&donationStatus)
	if err != nil {
		return nil, fmt.Errorf("gagal mengunci donasi: %w", err)
	}

	var active, delivered int

	err = tx.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status NOT IN ('delivered', 'cancelled')),
			COUNT(*) FILTER (WHERE status = 'delivered')
		FROM distributions
		WHERE donation_id = $1
	`, d.DonationID).Scan(&active, &delivered)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa distribusi donasi: %w", err)
	}

	if active == 0 {
		switch {
		case delivered == 0 && donationStatus == entity.DonationAllocated:
			err = setDonationStatus(ctx, tx, d.DonationID, entity.DonationAvailable, adminID)
		case delivered > 0 && donationStatus != entity.DonationDelivered:
			err = setDonationStatus(ctx, tx, d.DonationID, entity.DonationDelivered, adminID)
		}

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan pembatalan: %w", err)
	}

	return d, nil
}

// setDonationStatus mengubah status donasi dan mencatatnya ke donation_logs
// di dalam transaksi yang sedang berjalan.
func setDonationStatus(ctx context.Context, tx *sql.Tx, donationID int, status string, changedBy int) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE donations SET status = $1 WHERE id = $2`, status, donationID,
	)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status donasi: %w", err)
	}

	return insertDonationLog(ctx, tx, donationID, status, changedBy)
}
