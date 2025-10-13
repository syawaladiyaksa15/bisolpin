package usecase

import (
	"errors"
	"time"

	"main-service/internal/domain"
	"main-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Login(email, password string) (map[string]interface{}, error)
	Register(name, email, password, role string) (map[string]interface{}, error)
}

type userUsecase struct {
	repo       repository.UserRepository
	jwtSecret  string
	jwtExpHour int
}

// NewUserUsecase inisialisasi usecase dengan repo + secret jwt dari .env
func NewUserUsecase(repo repository.UserRepository, jwtSecret string, jwtExpHour int) UserUsecase {
	return &userUsecase{
		repo:       repo,
		jwtSecret:  jwtSecret,
		jwtExpHour: jwtExpHour,
	}
}

// -------------------- LOGIN --------------------

func (u *userUsecase) Login(email, password string) (map[string]interface{}, error) {
	if email == "" || password == "" {
		return nil, errors.New("email dan password wajib diisi")
	}

	user, err := u.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("email tidak ditemukan")
	}

	if user.IsActive == 0 {
		return nil, errors.New("akun tidak aktif")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("password salah")
	}

	tokenString, exp, err := u.generateToken(user)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	return map[string]interface{}{
		"token":      tokenString,
		"expires_at": exp.Format(time.RFC3339),
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil
}

// -------------------- REGISTER --------------------

func (u *userUsecase) Register(name, email, password, role string) (map[string]interface{}, error) {
	if name == "" || email == "" || password == "" || role == "" {
		return nil, errors.New("nama, email, password, dan role wajib diisi")
	}

	existing, _ := u.repo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashed),
		Role:     role,
		IsActive: 1,
	}

	if err := u.repo.CreateUser(user); err != nil {
		return nil, errors.New("gagal menyimpan user")
	}

	tokenString, exp, err := u.generateToken(user)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	return map[string]interface{}{
		"token":      tokenString,
		"expires_at": exp.Format(time.RFC3339),
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil
}

// -------------------- HELPER --------------------

func (u *userUsecase) generateToken(user *domain.User) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(u.jwtExpHour) * time.Hour)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, exp, nil
}
