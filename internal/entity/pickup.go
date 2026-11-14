package entity

import "time"

const (
	PickupAssigned  = "assigned"
	PickupPickedUp  = "picked_up"
	PickupDelivered = "delivered"
	PickupCancelled = "cancelled"
)

type Pickup struct {
	ID             int        `json:"id"`
	DistributionID int        `json:"distribution_id"`
	VolunteerID    int        `json:"volunteer_id"`
	Status         string     `json:"status"`
	PickedAt       *time.Time `json:"picked_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	Notes          string     `json:"notes"`
}

// PickupTask adalah tampilan gabungan (distribusi, donasi, mitra) untuk relawan.
// PickupID dan field setelahnya hanya terisi pada riwayat tugas relawan.
type PickupTask struct {
	PickupID          int        `json:"pickup_id,omitempty"`
	DistributionID    int        `json:"distribution_id"`
	DonationID        int        `json:"donation_id"`
	Title             string     `json:"title"`
	PortionsAllocated int        `json:"portions_allocated"`
	PickupAddress     string     `json:"pickup_address"`
	ExpiredAt         time.Time  `json:"expired_at"`
	PartnerName       string     `json:"partner_name"`
	PartnerAddress    string     `json:"partner_address"`
	Status            string     `json:"status"`
	PickedAt          *time.Time `json:"picked_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

// CanTransitionPickup menentukan perpindahan status pickup yang diizinkan:
// assigned -> picked_up -> delivered, atau assigned -> cancelled.
func CanTransitionPickup(from, to string) bool {
	switch from {
	case PickupAssigned:
		return to == PickupPickedUp || to == PickupCancelled
	case PickupPickedUp:
		return to == PickupDelivered
	}

	return false
}