package entity

import "time"

type DonationLog struct {
	ID         int       `json:"id"`
	DonationID int       `json:"donation_id"`
	Status     string    `json:"status"`
	ChangedBy  int       `json:"changed_by"`
	CreatedAt  time.Time `json:"created_at"`
}
