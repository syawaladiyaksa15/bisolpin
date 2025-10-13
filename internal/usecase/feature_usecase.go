package usecase

import (
	"errors"
	"main-service/internal/repository"
	"strings"
)

type FeatureUsecase interface {
	GetFeaturesByRole(role string) ([]repository.Feature, error)
	Create(name string, roles string, isActive *bool) (*repository.Feature, error)
	Update(id uint64, name string, roles string, isActive bool) (*repository.Feature, error)
	Delete(id uint64, role string) error
	GetDetail(id uint64) (*repository.Feature, error)
}

type featureUsecase struct {
	repo repository.FeatureRepository
}

func NewFeatureUsecase(r repository.FeatureRepository) FeatureUsecase {
	return &featureUsecase{repo: r}
}

func (u *featureUsecase) GetFeaturesByRole(role string) ([]repository.Feature, error) {
	return u.repo.GetByRole(role)
}

func (u *featureUsecase) Create(name string, roles string, isActive *bool) (*repository.Feature, error) {
	name = strings.TrimSpace(name)
	dup, err := u.repo.ExistsByName(name)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, errors.New("fitur dengan nama tersebut sudah ada")
	}

	if name == "" {
		return nil, errors.New("nama fitur tidak boleh kosong")
	}

	active := true
	if isActive != nil {
		active = *isActive
	}

	feature, err := u.repo.Create(name, roles, active)
	if err != nil {
		return nil, err
	}

	return feature, nil
}

func (u *featureUsecase) Update(id uint64, name string, roles string, isActive bool) (*repository.Feature, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama fitur wajib diisi")
	}

	roles = strings.TrimSpace(roles)
	if roles == "" {
		return nil, errors.New("roles wajib diisi")
	}

	dup, err := u.repo.ExistsByNameExceptID(id, name)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, errors.New("nama fitur sudah ada")
	}

	updated, err := u.repo.Update(id, name, roles, isActive)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (u *featureUsecase) Delete(id uint64, role string) error {
	if role != "admin" {
		return errors.New("akses ditolak, hanya admin yang dapat menghapus mata pelajaran")
	}

	return u.repo.Delete(id)
}

func (u *featureUsecase) GetDetail(id uint64) (*repository.Feature, error) {
	updated, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
