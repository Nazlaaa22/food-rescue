package repository

import (
	"context"
	"database/sql"
	"fmt"

	"food-rescue/internal/entity"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetStats mengumpulkan angka mentah dari tabel donations, distributions, partners, dan users.
// Perhitungan turunan (misalnya persentase) dilakukan di usecase.
func (r *StatsRepository) GetStats(ctx context.Context) (*entity.Stats, error) {
	stats := &entity.Stats{
		DonationsByStatus: map[string]int{
			entity.DonationAvailable: 0,
			entity.DonationAllocated: 0,
			entity.DonationPickedUp:  0,
			entity.DonationDelivered: 0,
			entity.DonationExpired:   0,
			entity.DonationCancelled: 0,
		},
		DistributionsByStatus: map[string]int{
			entity.DistributionAllocated: 0,
			entity.DistributionPickedUp:  0,
			entity.DistributionDelivered: 0,
			entity.DistributionCancelled: 0,
		},
	}

	if err := r.countDonations(ctx, stats); err != nil {
		return nil, err
	}

	if err := r.countDistributions(ctx, stats); err != nil {
		return nil, err
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active = TRUE)
		FROM partners
	`).Scan(&stats.TotalPartners, &stats.ActivePartners)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung mitra: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE role = 'donor'),
			COUNT(*) FILTER (WHERE role = 'volunteer'),
			COUNT(*) FILTER (WHERE role <> 'admin' AND COALESCE(is_verified, FALSE) = FALSE)
		FROM users
	`).Scan(&stats.TotalDonors, &stats.TotalVolunteers, &stats.UnverifiedUsers)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung pengguna: %w", err)
	}

	return stats, nil
}

func (r *StatsRepository) countDonations(ctx context.Context, stats *entity.Stats) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*), COALESCE(SUM(portions), 0)
		FROM donations
		GROUP BY status
	`)
	if err != nil {
		return fmt.Errorf("gagal menghitung donasi: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count, portions int

		if err := rows.Scan(&status, &count, &portions); err != nil {
			return fmt.Errorf("gagal membaca statistik donasi: %w", err)
		}

		stats.DonationsByStatus[status] = count
		stats.TotalDonations += count

		// Donasi yang dibatalkan donatur tidak dihitung sebagai porsi yang didonasikan.
		if status != entity.DonationCancelled {
			stats.TotalPortionsDonated += portions
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("terjadi kesalahan saat membaca statistik donasi: %w", err)
	}

	return nil
}

func (r *StatsRepository) countDistributions(ctx context.Context, stats *entity.Stats) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*), COALESCE(SUM(portions_allocated), 0)
		FROM distributions
		GROUP BY status
	`)
	if err != nil {
		return fmt.Errorf("gagal menghitung distribusi: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count, portions int

		if err := rows.Scan(&status, &count, &portions); err != nil {
			return fmt.Errorf("gagal membaca statistik distribusi: %w", err)
		}

		stats.DistributionsByStatus[status] = count
		stats.TotalDistributions += count

		if status != entity.DistributionCancelled {
			stats.PortionsAllocated += portions
		}

		// Porsi terselamatkan = porsi yang benar-benar sudah sampai di mitra.
		if status == entity.DistributionDelivered {
			stats.PortionsRescued += portions
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("terjadi kesalahan saat membaca statistik distribusi: %w", err)
	}

	return nil
}
