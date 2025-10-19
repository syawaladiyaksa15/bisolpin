package usecase

import (
	"errors"
	"main-service/internal/domain"
	"main-service/internal/repository"
)

type BimbelUsecase interface {
	Create(role string, userTutorID uint64, req *domain.Bimbel) error
	Update(role string, userTutorID uint64, req *domain.Bimbel) error
	Delete(role string, userTutorID uint64, id uint64) error
	FindByID(role string, userTutorID uint64, id uint64) (*domain.Bimbel, error)
	IsDuplicateName(name string, tutorID uint64) (bool, error)
}

type bimbelUsecase struct {
	repo repository.BimbelRepository
}

func NewBimbelUsecase(r repository.BimbelRepository) BimbelUsecase {
	return &bimbelUsecase{repo: r}
}

func (u *bimbelUsecase) Create(role string, userTutorID uint64, req *domain.Bimbel) error {
	if req.SubjectID == 0 || req.Thumbnail == "" || req.Deskripsi == "" || req.Harga <= 0 {
		return errors.New("all required fields must be filled")
	}

	if role == "tutor" {
		req.TutorID = userTutorID
	} else if role != "admin" {
		return errors.New("forbidden")
	}

	exists, _ := u.repo.ExistsDuplicate(req.Name, req.FeatureID, req.SubjectID, nil)
	if exists {
		return errors.New("duplicate bimbel name for this feature and subject")
	}

	if !req.IsActive {
		req.IsActive = true
	}

	return u.repo.Create(req)
}

func (u *bimbelUsecase) Update(role string, userTutorID uint64, req *domain.Bimbel) error {
	existing, err := u.repo.FindByID(req.ID)
	if err != nil {
		return err
	}

	if role == "tutor" && existing.TutorID != userTutorID {
		return errors.New("unauthorized")
	}

	exists, _ := u.repo.ExistsDuplicate(req.Name, req.FeatureID, req.SubjectID, &req.ID)
	if exists {
		return errors.New("duplicate bimbel name for this feature and subject")
	}

	return u.repo.Update(req)
}

func (u *bimbelUsecase) Delete(role string, userTutorID uint64, id uint64) error {
	b, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}

	if role == "tutor" && b.TutorID != userTutorID {
		return errors.New("unauthorized")
	}

	return u.repo.Delete(id)
}

func (u *bimbelUsecase) FindByID(role string, userTutorID uint64, id uint64) (*domain.Bimbel, error) {
	b, err := u.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if role == "tutor" && b.TutorID != userTutorID {
		return nil, errors.New("unauthorized")
	}
	return b, nil
}

func (u *bimbelUsecase) IsDuplicateName(name string, tutorID uint64) (bool, error) {
	return u.repo.ExistsByNameAndTutor(name, tutorID)
}
