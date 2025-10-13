package repository

import (
	"database/sql"
	"errors"
)

type Matpel struct {
	ID        uint64  `json:"id"`
	FeatureID uint64  `json:"feature_id"`
	Name      string  `json:"name"`
	Deskripsi *string `json:"deskripsi,omitempty"`
	IsActive  bool    `json:"is_active"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

type MatpelRepository interface {
	GetByFeature(featureId uint64) ([]Matpel, error)
	Create(featureID uint64, name string, deskripsi *string, isActive bool) (*Matpel, error)
	ExistsByNameAndFeatureID(name string, featureID uint64) (bool, error)
	Update(id uint64, featureID uint64, name string, deskripsi *string, isActive bool) (*Matpel, error)
	ExistsByNameAndFeatureIDExceptID(id uint64, featureID uint64, name string) (bool, error)
	Delete(id uint64) error
	GetByID(id uint64) (*Matpel, error)
}

type matpelRepository struct {
	db *sql.DB
}

func NewMatpelRepository(db *sql.DB) MatpelRepository {
	return &matpelRepository{db: db}
}

func (r *matpelRepository) GetByFeature(featureId uint64) ([]Matpel, error) {
	query := `
		SELECT id, feature_id, name, deskripsi, is_active, created_at, updated_at
		FROM subjects
		WHERE is_active = 1
		  AND feature_id = ?
	`

	rows, err := r.db.Query(query, featureId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matpels []Matpel
	for rows.Next() {
		var f Matpel
		if err := rows.Scan(&f.ID, &f.FeatureID, &f.Name, &f.Deskripsi, &f.IsActive, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		matpels = append(matpels, f)
	}
	return matpels, nil
}

func (r *matpelRepository) Create(featureID uint64, name string, deskripsi *string, isActive bool) (*Matpel, error) {
	query := `
		INSERT INTO subjects (feature_id, name, deskripsi, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	res, err := r.db.Exec(query, featureID, name, deskripsi, isActive)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	subject := &Matpel{
		ID:        uint64(id),
		FeatureID: featureID,
		Name:      name,
		Deskripsi: deskripsi,
		IsActive:  isActive,
	}
	return subject, nil
}

func (r *matpelRepository) ExistsByNameAndFeatureID(name string, featureID uint64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM subjects 
			WHERE feature_id = ? AND LOWER(TRIM(name)) = LOWER(TRIM(?))
		)
	`
	err := r.db.QueryRow(query, featureID, name).Scan(&exists)
	return exists, err
}

func (r *matpelRepository) Update(id uint64, featureID uint64, name string, deskripsi *string, isActive bool) (*Matpel, error) {
	_, err := r.db.Exec(`
		UPDATE subjects
		SET feature_id = ?, name = ?, deskripsi = ?, is_active = ?, updated_at = NOW()
		WHERE id = ?`, featureID, name, deskripsi, isActive, id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *matpelRepository) GetByID(id uint64) (*Matpel, error) {
	var m Matpel
	err := r.db.QueryRow(`
		SELECT id, feature_id, name, deskripsi, is_active, created_at, updated_at
		FROM subjects WHERE id = ?`, id).
		Scan(&m.ID, &m.FeatureID, &m.Name, &m.Deskripsi, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("matpel not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *matpelRepository) ExistsByNameAndFeatureIDExceptID(id uint64, featureID uint64, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM subjects 
		WHERE feature_id = ? AND name = ? AND id <> ?`, featureID, name, id).Scan(&count)
	return count > 0, err
}

func (r *matpelRepository) Delete(id uint64) error {
	query := `DELETE FROM subjects WHERE id = ?`
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
