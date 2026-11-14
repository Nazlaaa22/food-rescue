package usecase

import (
	"context"
	"math"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type StatsUsecase struct {
	statsRepo *repository.StatsRepository
}

func NewStatsUsecase(
	statsRepo *repository.StatsRepository,
) *StatsUsecase {
	return &StatsUsecase{statsRepo: statsRepo}
}

// Get mengembalikan ringkasan statistik sistem untuk Admin, termasuk
// persentase porsi terselamatkan dibanding total porsi yang didonasikan.
func (u *StatsUsecase) Get(ctx context.Context) (*entity.Stats, error) {
	stats, err := u.statsRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	stats.RescueRatePercent = rescueRate(stats.PortionsRescued, stats.TotalPortionsDonated)

	return stats, nil
}

// rescueRate menghitung persentase dengan dua angka di belakang koma.
func rescueRate(rescued, donated int) float64 {
	if donated <= 0 {
		return 0
	}

	rate := float64(rescued) / float64(donated) * 100

	return math.Round(rate*100) / 100
}
