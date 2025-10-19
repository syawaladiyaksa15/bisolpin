package repository

import (
	"database/sql"
	"errors"
	"main-service/internal/domain"
	"time"
)

type BimbelRepository interface {
	Create(b *domain.Bimbel) error
	Update(b *domain.Bimbel) error
	Delete(id uint64) error
	FindByID(id uint64) (*domain.Bimbel, error)
	ExistsDuplicate(name string, featureID, subjectID uint64, excludeID *uint64) (bool, error)
	FindByTutor(id uint64) ([]domain.Bimbel, error)
	ExistsByNameAndTutor(name string, tutorID uint64) (bool, error)
}

type bimbelRepository struct {
	db *sql.DB
}

func NewBimbelRepository(db *sql.DB) BimbelRepository {
	return &bimbelRepository{db}
}

func (r *bimbelRepository) ExistsDuplicate(name string, featureID, subjectID uint64, excludeID *uint64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bimbels 
		WHERE name = ? AND feature_id = ? AND subject_id = ? AND deleted_at IS NULL
	`
	args := []interface{}{name, featureID, subjectID}
	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *bimbelRepository) Create(b *domain.Bimbel) error {
	query := `
		INSERT INTO bimbels (tutor_id, feature_id, subject_id, name, limit_peserta, is_active, thumbnail, deskripsi, harga, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	res, err := r.db.Exec(query,
		b.TutorID, b.FeatureID, b.SubjectID,
		b.Name, b.LimitPeserta, b.IsActive,
		b.Thumbnail, b.Deskripsi, b.Harga,
	)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	b.ID = uint64(id)
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return nil
}

func (r *bimbelRepository) Update(b *domain.Bimbel) error {
	query := `
		UPDATE bimbels SET feature_id=?, subject_id=?, name=?, limit_peserta=?, is_active=?, thumbnail=?, deskripsi=?, harga=?, updated_at=NOW()
		WHERE id=? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, b.FeatureID, b.SubjectID, b.Name, b.LimitPeserta, b.IsActive, b.Thumbnail, b.Deskripsi, b.Harga, b.ID)
	return err
}

func (r *bimbelRepository) Delete(id uint64) error {
	_, err := r.db.Exec(`UPDATE bimbels SET deleted_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *bimbelRepository) FindByID(id uint64) (*domain.Bimbel, error) {
	query := `
		SELECT id, tutor_id, feature_id, subject_id, name, limit_peserta, is_active, thumbnail, deskripsi, harga, created_at, updated_at
		FROM bimbels WHERE id = ? AND deleted_at IS NULL
	`
	var b domain.Bimbel
	err := r.db.QueryRow(query, id).Scan(
		&b.ID, &b.TutorID, &b.FeatureID, &b.SubjectID, &b.Name, &b.LimitPeserta,
		&b.IsActive, &b.Thumbnail, &b.Deskripsi, &b.Harga, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("bimbel not found")
		}
		return nil, err
	}
	return &b, nil
}

func (r *bimbelRepository) FindByTutor(tutorID uint64) ([]domain.Bimbel, error) {
	query := `
		SELECT id, tutor_id, feature_id, subject_id, name, limit_peserta, is_active, thumbnail, deskripsi, harga, created_at, updated_at
		FROM bimbels WHERE tutor_id = ? AND deleted_at IS NULL
	`
	rows, err := r.db.Query(query, tutorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Bimbel
	for rows.Next() {
		var b domain.Bimbel
		rows.Scan(&b.ID, &b.TutorID, &b.FeatureID, &b.SubjectID, &b.Name, &b.LimitPeserta,
			&b.IsActive, &b.Thumbnail, &b.Deskripsi, &b.Harga, &b.CreatedAt, &b.UpdatedAt)
		result = append(result, b)
	}
	return result, nil
}

func (r *bimbelRepository) ExistsByNameAndTutor(name string, tutorID uint64) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM bimbels 
		WHERE name = ? AND tutor_id = ? AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(query, name, tutorID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
