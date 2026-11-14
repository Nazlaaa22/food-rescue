package repository

import (
	"context"
	"database/sql"
	"fmt"

	"food-rescue/internal/entity"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userSelect = `
	SELECT id, name, email, password_hash, COALESCE(phone, ''), role,
	       COALESCE(is_verified, FALSE), created_at
	FROM users
`

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			name, email, password_hash, phone, role, is_verified
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.Role,
		user.IsVerified,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		// 23505 = unique violation: email sudah dipakai (misalnya dua register bersamaan).
		if hasPgErrorCode(err, "23505") {
			return entity.NewDomainError(entity.ErrConflict, "email sudah terdaftar")
		}

		return fmt.Errorf("gagal membuat user: %w", err)
	}

	return nil
}

func scanUser(row rowScanner) (*entity.User, error) {
	user := &entity.User{}

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByEmail mengembalikan sql.ErrNoRows kalau email belum terdaftar.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, userSelect+" WHERE email = $1", email))
}

// FindByID mengembalikan sql.ErrNoRows kalau pengguna tidak ditemukan.
func (r *UserRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, userSelect+" WHERE id = $1", id))
}

// FindAll dipakai admin. role kosong berarti semua role,
// verified nil berarti semua status verifikasi.
func (r *UserRepository) FindAll(ctx context.Context, role string, verified *bool) ([]entity.User, error) {
	query := userSelect + `
		WHERE ($1::text = '' OR role = $1::text)
		  AND ($2::boolean IS NULL OR COALESCE(is_verified, FALSE) = $2::boolean)
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, role, verified)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pengguna: %w", err)
	}
	defer rows.Close()

	users := []entity.User{}

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca data pengguna: %w", err)
		}

		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca data pengguna: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, user *entity.User) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = $1, phone = $2 WHERE id = $3`,
		user.Name, user.Phone, user.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal memperbarui profil: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal memeriksa hasil pembaruan profil: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepository) UpdateVerificationStatus(
	ctx context.Context,
	userID int,
	isVerified bool,
) error {
	// Akun admin tidak ikut diverifikasi lewat endpoint ini.
	query := `
		UPDATE users
		SET is_verified = $1
		WHERE id = $2 AND role <> 'admin'
	`

	result, err := r.db.ExecContext(ctx, query, isVerified, userID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status verifikasi: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal memeriksa hasil verifikasi: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
