package entity

import "time"

const (
	DonationAvailable = "available"
	DonationAllocated = "allocated"
	DonationPickedUp  = "picked_up"
	DonationDelivered = "delivered"
	DonationExpired   = "expired"
	DonationCancelled = "cancelled"
)

type Donation struct {
	ID            int       `json:"id"`
	DonorID       int       `json:"donor_id"`
	CategoryID    int       `json:"category_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Portions      int       `json:"portions"`
	ExpiredAt     time.Time `json:"expired_at"`
	PickupAddress string    `json:"pickup_address"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

func IsValidDonationStatus(status string) bool {
	switch status {
	case DonationAvailable, DonationAllocated, DonationPickedUp,
		DonationDelivered, DonationExpired, DonationCancelled:
		return true
	}

	return false
}