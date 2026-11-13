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
		return fmt.Errorf("gagal membuat user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}

	query := `
		SELECT id, name, email, password_hash, phone, role, is_verified, created_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
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

func (r *UserRepository) UpdateVerificationStatus(
	ctx context.Context,
	userID int,
	isVerified bool,
) error {
	query := `
		UPDATE users
		SET is_verified = $1
		WHERE id = $2
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
