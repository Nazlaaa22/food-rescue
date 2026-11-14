package entity

import "errors"

// Error dasar yang dipakai lintas layer.
// Handler HTTP memetakan error ini ke status code yang sesuai.
var (
	ErrInvalid   = errors.New("input tidak valid")
	ErrForbidden = errors.New("akses ditolak")
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrConflict  = errors.New("data bertentangan dengan kondisi saat ini")
)

// DomainError membawa pesan yang aman ditampilkan ke client.
type DomainError struct {
	Kind    error
	Message string
}

func (e *DomainError) Error() string { return e.Message }

func (e *DomainError) Unwrap() error { return e.Kind }

func NewDomainError(kind error, message string) error {
	return &DomainError{Kind: kind, Message: message}
}