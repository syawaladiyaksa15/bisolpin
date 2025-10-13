package repository

import (
	"database/sql"
	"errors"
	"main-service/internal/domain"
)

type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	CreateUser(user *domain.User) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password, role, is_active
		FROM users
		WHERE email = ? AND is_active = 1 AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.IsActive)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (name, email, password, role, is_active, created_at)
		VALUES (?, ?, ?, ?, 1, NOW())
	`
	result, err := r.db.Exec(query, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.New("failed to get inserted ID")
	}
	user.ID = uint64(id)
	return nil
}
