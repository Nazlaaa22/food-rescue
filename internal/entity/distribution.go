package entity

import "time"

const (
	DistributionAllocated = "allocated"
	DistributionPickedUp  = "picked_up"
	DistributionDelivered = "delivered"
	DistributionCancelled = "cancelled"
)

// Distribution adalah alokasi sebagian (atau seluruh) porsi donasi ke satu mitra.
// Satu donasi boleh punya banyak distribusi ke mitra yang berbeda.
type Distribution struct {
	ID                int       `json:"id"`
	DonationID        int       `json:"donation_id"`
	PartnerID         int       `json:"partner_id"`
	AdminID           int       `json:"admin_id"`
	PortionsAllocated int       `json:"portions_allocated"`
	Status            string    `json:"status"`
	AllocatedAt       time.Time `json:"allocated_at"`
}

// DistributionDetail adalah tampilan gabungan (distribusi, donasi, mitra, pickup)
// untuk Admin memantau penyaluran. Field pickup kosong kalau belum ada relawan.
type DistributionDetail struct {
	Distribution
	DonationTitle string  `json:"donation_title"`
	PartnerName   string  `json:"partner_name"`
	VolunteerID   *int    `json:"volunteer_id,omitempty"`
	PickupStatus  *string `json:"pickup_status,omitempty"`
}

// DistributionFilter dipakai Admin untuk menyaring daftar distribusi.
// Nilai kosong atau nol berarti tidak disaring.
type DistributionFilter struct {
	Status     string
	DonationID int
	PartnerID  int
}

func IsValidDistributionStatus(status string) bool {
	switch status {
	case DistributionAllocated, DistributionPickedUp,
		DistributionDelivered, DistributionCancelled:
		return true
	}

	return false
}
