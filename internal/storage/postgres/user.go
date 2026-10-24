package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrConflict = errors.New("conflict")

// CreateUser - создает пользователя
func (s *Storage) CreateUser(ctx context.Context, user models.User) error {
	const fn = "storage.postgres.user.CreateUser"

	_, err := s.db.Exec(ctx,
		`INSERT INTO users (name, email) VALUES ($1, $2)`,
		user.Name, user.Email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w: email=%s", fn, ErrConflict, user.Email)
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// GetUserByEmail - получает пользователя по email
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const fn = "storage.postgres.user.GetUserByEmail"

	var user models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, name, email FROM users WHERE email = $1`,
		email).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w: email=%s", fn, ErrNotFound, email)
		}
		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return user, nil
}

// GetAllUsers - получает всех пользователей
func (s *Storage) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const fn = "storage.postgres.user.GetAllUsers"

	rows, err := s.db.Query(ctx, `SELECT id, name, email FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return users, nil
}
