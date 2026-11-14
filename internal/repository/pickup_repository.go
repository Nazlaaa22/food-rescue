package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"food-rescue/internal/entity"
)

type PickupRepository struct {
	db *sql.DB
}

func NewPickupRepository(db *sql.DB) *PickupRepository {
	return &PickupRepository{db: db}
}

// FindAvailable mengambil distribusi berstatus allocated yang belum diambil relawan
// dan donasinya belum melewati batas waktu layak konsumsi.
func (r *PickupRepository) FindAvailable(ctx context.Context) ([]entity.PickupTask, error) {
	query := `
		SELECT d.id, dn.id, dn.title, d.portions_allocated, dn.pickup_address, dn.expired_at,
		       pt.name, pt.address, d.status
		FROM distributions d
		JOIN donations dn ON dn.id = d.donation_id
		JOIN partners pt ON pt.id = d.partner_id
		WHERE d.status = 'allocated'
		  AND dn.expired_at > (NOW() AT TIME ZONE 'UTC')
		  AND NOT EXISTS (
			SELECT 1 FROM pickups p
			WHERE p.distribution_id = d.id AND p.status <> 'cancelled'
		  )
		ORDER BY dn.expired_at ASC, d.id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil tugas tersedia: %w", err)
	}
	defer rows.Close()

	return scanPickupTasks(rows, false)
}

// FindByVolunteerID mengambil riwayat tugas milik satu relawan.
func (r *PickupRepository) FindByVolunteerID(ctx context.Context, volunteerID int) ([]entity.PickupTask, error) {
	query := `
		SELECT d.id, dn.id, dn.title, d.portions_allocated, dn.pickup_address, dn.expired_at,
		       pt.name, pt.address, p.status,
		       p.id, p.picked_at, p.delivered_at, COALESCE(p.notes, '')
		FROM pickups p
		JOIN distributions d ON d.id = p.distribution_id
		JOIN donations dn ON dn.id = d.donation_id
		JOIN partners pt ON pt.id = d.partner_id
		WHERE p.volunteer_id = $1
		ORDER BY p.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, volunteerID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat tugas: %w", err)
	}
	defer rows.Close()

	return scanPickupTasks(rows, true)
}

func scanPickupTasks(rows *sql.Rows, withPickup bool) ([]entity.PickupTask, error) {
	tasks := []entity.PickupTask{}

	for rows.Next() {
		var task entity.PickupTask
		var pickedAt, deliveredAt sql.NullTime

		dest := []any{
			&task.DistributionID,
			&task.DonationID,
			&task.Title,
			&task.PortionsAllocated,
			&task.PickupAddress,
			&task.ExpiredAt,
			&task.PartnerName,
			&task.PartnerAddress,
			&task.Status,
		}

		if withPickup {
			dest = append(dest, &task.PickupID, &pickedAt, &deliveredAt, &task.Notes)
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("gagal membaca tugas: %w", err)
		}

		if pickedAt.Valid {
			task.PickedAt = &pickedAt.Time
		}

		if deliveredAt.Valid {
			task.DeliveredAt = &deliveredAt.Time
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca tugas: %w", err)
	}

	return tasks, nil
}

// Claim membuat pickup baru (status assigned) untuk satu distribusi.
// Baris distribusi dikunci supaya dua relawan tidak bisa mengambil tugas yang sama.
func (r *PickupRepository) Claim(ctx context.Context, volunteerID, distributionID int) (*entity.Pickup, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	var distributionStatus string
	var notExpired bool

	err = tx.QueryRowContext(ctx, `
		SELECT d.status, dn.expired_at > (NOW() AT TIME ZONE 'UTC')
		FROM distributions d
		JOIN donations dn ON dn.id = d.donation_id
		WHERE d.id = $1
		FOR UPDATE OF d
	`, distributionID).Scan(&distributionStatus, &notExpired)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "tugas tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa tugas: %w", err)
	}

	if distributionStatus != "allocated" {
		return nil, entity.NewDomainError(entity.ErrConflict, "tugas sudah tidak tersedia")
	}

	if !notExpired {
		return nil, entity.NewDomainError(entity.ErrConflict, "donasi sudah melewati batas waktu layak konsumsi")
	}

	var alreadyTaken bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pickups
			WHERE distribution_id = $1 AND status <> 'cancelled'
		)
	`, distributionID).Scan(&alreadyTaken)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa pickup: %w", err)
	}

	if alreadyTaken {
		return nil, entity.NewDomainError(entity.ErrConflict, "tugas sudah diambil relawan lain")
	}

	pickup := &entity.Pickup{
		DistributionID: distributionID,
		VolunteerID:    volunteerID,
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO pickups (distribution_id, volunteer_id)
		VALUES ($1, $2)
		RETURNING id, status
	`, distributionID, volunteerID).Scan(&pickup.ID, &pickup.Status)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil tugas: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan pickup: %w", err)
	}

	return pickup, nil
}

// UpdateStatus mengubah status pickup milik relawan, lalu menyelaraskan status
// distribusi dan donasi (beserta donation_logs) dalam satu transaksi.
func (r *PickupRepository) UpdateStatus(
	ctx context.Context,
	volunteerID, pickupID int,
	newStatus, notes string,
) (*entity.Pickup, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	pickup := &entity.Pickup{ID: pickupID}
	var donationID int

	err = tx.QueryRowContext(ctx, `
		SELECT p.distribution_id, p.volunteer_id, p.status, d.donation_id
		FROM pickups p
		JOIN distributions d ON d.id = p.distribution_id
		WHERE p.id = $1
		FOR UPDATE OF p
	`, pickupID).Scan(&pickup.DistributionID, &pickup.VolunteerID, &pickup.Status, &donationID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewDomainError(entity.ErrNotFound, "tugas tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil pickup: %w", err)
	}

	if pickup.VolunteerID != volunteerID {
		return nil, entity.NewDomainError(entity.ErrForbidden, "tugas ini bukan milik Anda")
	}

	if !entity.CanTransitionPickup(pickup.Status, newStatus) {
		return nil, entity.NewDomainError(
			entity.ErrConflict,
			fmt.Sprintf("status tidak dapat diubah dari %s ke %s", pickup.Status, newStatus),
		)
	}

	var pickedAt, deliveredAt sql.NullTime

	err = tx.QueryRowContext(ctx, `
		UPDATE pickups
		SET status = $1::text,
		    picked_at = CASE WHEN $1::text = 'picked_up'
		                     THEN (NOW() AT TIME ZONE 'UTC') ELSE picked_at END,
		    delivered_at = CASE WHEN $1::text = 'delivered'
		                        THEN (NOW() AT TIME ZONE 'UTC') ELSE delivered_at END,
		    notes = COALESCE(NULLIF($2::text, ''), notes)
		WHERE id = $3
		RETURNING picked_at, delivered_at, COALESCE(notes, '')
	`, newStatus, notes, pickupID).Scan(&pickedAt, &deliveredAt, &pickup.Notes)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui pickup: %w", err)
	}

	pickup.Status = newStatus

	if pickedAt.Valid {
		pickup.PickedAt = &pickedAt.Time
	}

	if deliveredAt.Valid {
		pickup.DeliveredAt = &deliveredAt.Time
	}

	// Membatalkan pickup tidak mengubah distribusi: statusnya tetap allocated
	// sehingga tugas muncul lagi di daftar tugas tersedia.
	if newStatus != entity.PickupCancelled {
		err = syncDistributionAndDonation(ctx, tx, pickup.DistributionID, donationID, volunteerID, newStatus)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan perubahan status: %w", err)
	}

	return pickup, nil
}

// syncDistributionAndDonation menyelaraskan status distribusi dan donasi.
//   - picked_up: donasi menjadi picked_up saat pickup pertamanya diambil.
//   - delivered: donasi menjadi delivered setelah semua distribusinya
//     delivered atau cancelled.
func syncDistributionAndDonation(
	ctx context.Context,
	tx *sql.Tx,
	distributionID, donationID, changedBy int,
	newStatus string,
) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE distributions SET status = $1::text WHERE id = $2`,
		newStatus, distributionID,
	)
	if err != nil {
		return fmt.Errorf("gagal memperbarui distribusi: %w", err)
	}

	switch newStatus {
	case entity.PickupPickedUp:
		res, err := tx.ExecContext(ctx, `
			UPDATE donations SET status = 'picked_up'
			WHERE id = $1 AND status = 'allocated'
		`, donationID)
		if err != nil {
			return fmt.Errorf("gagal memperbarui donasi: %w", err)
		}

		changed, err := rowsChanged(res)
		if err != nil {
			return err
		}

		if changed {
			return insertDonationLog(ctx, tx, donationID, entity.DonationPickedUp, changedBy)
		}

	case entity.PickupDelivered:
		// Kunci baris donasi supaya dua pengantaran bersamaan tidak saling melewatkan
		// pemeriksaan "semua distribusi sudah selesai".
		var lockedID int

		err := tx.QueryRowContext(ctx,
			`SELECT id FROM donations WHERE id = $1 FOR UPDATE`, donationID,
		).Scan(&lockedID)
		if err != nil {
			return fmt.Errorf("gagal mengunci donasi: %w", err)
		}

		var remaining int

		err = tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM distributions
			WHERE donation_id = $1 AND status NOT IN ('delivered', 'cancelled')
		`, donationID).Scan(&remaining)
		if err != nil {
			return fmt.Errorf("gagal memeriksa distribusi: %w", err)
		}

		if remaining > 0 {
			return nil
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE donations SET status = 'delivered'
			WHERE id = $1 AND status <> 'delivered'
		`, donationID)
		if err != nil {
			return fmt.Errorf("gagal memperbarui donasi: %w", err)
		}

		changed, err := rowsChanged(res)
		if err != nil {
			return err
		}

		if changed {
			return insertDonationLog(ctx, tx, donationID, entity.DonationDelivered, changedBy)
		}
	}

	return nil
}

func rowsChanged(res sql.Result) (bool, error) {
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa hasil pembaruan: %w", err)
	}

	return n > 0, nil
}