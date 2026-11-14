package entity

// Stats adalah ringkasan aktivitas sistem untuk dashboard Admin.
type Stats struct {
	TotalDonations        int            `json:"total_donations"`
	DonationsByStatus     map[string]int `json:"donations_by_status"`
	TotalPortionsDonated  int            `json:"total_portions_donated"`
	TotalDistributions    int            `json:"total_distributions"`
	DistributionsByStatus map[string]int `json:"distributions_by_status"`
	PortionsAllocated     int            `json:"portions_allocated"`
	PortionsRescued       int            `json:"portions_rescued"`
	RescueRatePercent     float64        `json:"rescue_rate_percent"`
	TotalPartners         int            `json:"total_partners"`
	ActivePartners        int            `json:"active_partners"`
	TotalDonors           int            `json:"total_donors"`
	TotalVolunteers       int            `json:"total_volunteers"`
	UnverifiedUsers       int            `json:"unverified_users"`
}
