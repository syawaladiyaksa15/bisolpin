package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

type Feature struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	IsActive  bool    `json:"is_active"`
	Roles     string  `json:"roles"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

type FeatureRepository interface {
	GetByRole(role string) ([]Feature, error)
	ExistsByID(id uint64) (bool, error)
	ExistsByName(name string) (bool, error)
	Create(name string, roles string, isActive bool) (*Feature, error)
	ExistsByNameExceptID(id uint64, name string) (bool, error)
	Update(id uint64, name string, roles string, isActive bool) (*Feature, error)
	GetByID(id uint64) (*Feature, error)
	Delete(id uint64) error
}

type featureRepository struct {
	db *sql.DB
}

func NewFeatureRepository(db *sql.DB) FeatureRepository {
	return &featureRepository{db: db}
}

func (r *featureRepository) GetByRole(role string) ([]Feature, error) {
	query := `
		SELECT id, name, is_active, roles, created_at, updated_at
		FROM features
		WHERE is_active = 1
		  AND (
			roles LIKE ? OR
			roles LIKE ? OR
			roles LIKE ? OR
			roles = ?
		  )
	`
	rolePattern := fmt.Sprintf("%%,%s,%%", role)
	rolePatternStart := fmt.Sprintf("%s,%%", role)
	rolePatternEnd := fmt.Sprintf("%%,%s", role)

	rows, err := r.db.Query(query, rolePattern, rolePatternStart, rolePatternEnd, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []Feature
	for rows.Next() {
		var f Feature
		if err := rows.Scan(&f.ID, &f.Name, &f.IsActive, &f.Roles, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, nil
}

func (r *featureRepository) ExistsByID(id uint64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM features WHERE id = ? AND is_active = 1)`
	err := r.db.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (r *featureRepository) ExistsByName(name string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM features 
			WHERE LOWER(TRIM(name)) = LOWER(TRIM(?))
		)
	`
	err := r.db.QueryRow(query, name).Scan(&exists)
	return exists, err
}

func (r *featureRepository) Create(name string, roles string, isActive bool) (*Feature, error) {
	query := `
		INSERT INTO features (name, is_active, roles, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	res, err := r.db.Exec(query, name, isActive, roles)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	feature := &Feature{
		ID:   uint64(id),
		Name: name,
	}
	return feature, nil
}

func (r *featureRepository) ExistsByNameExceptID(id uint64, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM features 
		WHERE name = ? AND id <> ?`, name, id).Scan(&count)
	return count > 0, err
}

func (r *featureRepository) Update(id uint64, name string, roles string, isActive bool) (*Feature, error) {
	_, err := r.db.Exec(`
		UPDATE features
		SET name = ?, is_active = ?, roles = ?, updated_at = NOW()
		WHERE id = ?`, name, isActive, roles, id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *featureRepository) GetByID(id uint64) (*Feature, error) {
	var f Feature
	err := r.db.QueryRow(`
		SELECT id, name, is_active, roles, created_at, updated_at
		FROM features WHERE id = ?`, id).
		Scan(&f.ID, &f.Name, &f.IsActive, &f.Roles, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("feature not found")
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *featureRepository) Delete(id uint64) error {
	query := `DELETE FROM features WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("data tidak ditemukan")
	}

	return nil
}
