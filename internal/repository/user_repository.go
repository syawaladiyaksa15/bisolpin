package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"main-service/internal/domain"
)

type User struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	TutorID   *uint64
	PesertaID *uint64
}

type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	CreateUser(user *domain.User) error
	FindTutorIDByUserID(userID uint64) (*domain.User, error)
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
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var tutorID, pesertaID sql.NullInt64

	// ==== 1️⃣ Buat relasi tutor/peserta bila diperlukan ====
	if user.Role == "tutor" {
		queryTutor := `INSERT INTO tutors (is_active, created_at) VALUES (1, NOW())`
		res, err := tx.Exec(queryTutor)
		if err != nil {
			return fmt.Errorf("gagal insert tutor: %v", err)
		}
		lastID, _ := res.LastInsertId()
		tutorID = sql.NullInt64{Int64: lastID, Valid: true}
	} else if user.Role == "peserta" {
		queryPeserta := `INSERT INTO pesertas (is_active, created_at) VALUES (1, NOW())`
		res, err := tx.Exec(queryPeserta)
		if err != nil {
			return fmt.Errorf("gagal insert peserta: %v", err)
		}
		lastID, _ := res.LastInsertId()
		pesertaID = sql.NullInt64{Int64: lastID, Valid: true}
	}

	// ==== 2️⃣ Insert ke users ====
	query := `
		INSERT INTO users (name, email, password, role, tutor_id, peserta_id, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, NOW())
	`
	res, err := tx.Exec(query,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		tutorID,
		pesertaID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal insert user: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal ambil user id: %v", err)
	}
	user.ID = uint64(lastID)

	// ==== 3️⃣ Commit transaksi ====
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal commit transaksi: %v", err)
	}

	// Set nilai tutor/peserta ID ke struct user
	if tutorID.Valid {
		user.TutorID = &[]uint64{uint64(tutorID.Int64)}[0]
	}
	if pesertaID.Valid {
		user.PesertaID = &[]uint64{uint64(pesertaID.Int64)}[0]
	}

	return nil
}

// func (r *userRepository) FindTutorIDByUserID(userID uint64) (*uint64, error) {
// 	query := `
// 		SELECT tutor_id
// 		FROM users
// 		WHERE id = ? AND role = 'tutor' AND deleted_at IS NULL
// 	`
// 	row := r.db.QueryRow(query, userID)

// 	var tutorID sql.NullInt64
// 	if err := row.Scan(&tutorID); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, errors.New("user tidak ditemukan atau bukan tutor")
// 		}
// 		return nil, err
// 	}

// 	if !tutorID.Valid {
// 		return nil, errors.New("tutor_id kosong")
// 	}

// 	id := uint64(tutorID.Int64)
// 	return &id, nil
// }

func (r *userRepository) FindTutorIDByUserID(userID uint64) (*domain.User, error) {
	query := `
		SELECT id, name, email, role, tutor_id, peserta_id
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, userID)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.TutorID, &user.PesertaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
	}

	return &user, nil
}
