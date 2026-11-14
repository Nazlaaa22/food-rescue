package entity

// Partner adalah mitra penerima donasi (panti asuhan, panti jompo, komunitas, dll).
// Mitra bukan user yang bisa login, melainkan data yang dikelola Admin.
type Partner struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"is_active"`
}
