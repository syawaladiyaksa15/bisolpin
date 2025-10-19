package domain

import "time"

type Bimbel struct {
	ID           uint64    `json:"id"`
	TutorID      uint64    `json:"tutor_id"`
	FeatureID    uint64    `json:"feature_id"`
	SubjectID    uint64    `json:"subject_id"`
	Name         string    `json:"name"`
	LimitPeserta int       `json:"limit_peserta"`
	IsActive     bool      `json:"is_active"`
	Thumbnail    string    `json:"thumbnail"`
	Deskripsi    string    `json:"deskripsi"`
	Harga        float64   `json:"harga"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
