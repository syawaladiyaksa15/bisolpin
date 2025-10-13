package usecase

import (
	"errors"
	"main-service/internal/repository"
	"strings"
)

type MatpelUsecase interface {
	GetMatpelByFeature(featureId uint64) ([]repository.Matpel, error)
	Create(featureID uint64, name string, deskripsi *string, isActive *bool) (*repository.Matpel, error)
	Update(id uint64, featureID uint64, name string, deskripsi *string, isActive bool) (*repository.Matpel, error)
	Delete(id uint64, role string) error
	GetDetail(id uint64) (*repository.Matpel, error)
}

type matpelUsecase struct {
	matpelRepo  repository.MatpelRepository
	featureRepo repository.FeatureRepository
}

func NewMatpelUsecase(subjectRepo repository.MatpelRepository, featureRepo repository.FeatureRepository) MatpelUsecase {
	return &matpelUsecase{
		matpelRepo:  subjectRepo,
		featureRepo: featureRepo,
	}
}

func (u *matpelUsecase) GetMatpelByFeature(featureId uint64) ([]repository.Matpel, error) {
	return u.matpelRepo.GetByFeature(featureId)
}

func (u *matpelUsecase) Create(featureID uint64, name string, deskripsi *string, isActive *bool) (*repository.Matpel, error) {
	// Check: Feature ID valid?
	exists, err := u.featureRepo.ExistsByID(featureID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("feature_id tidak ditemukan")
	}

	name = strings.TrimSpace(name)
	// Check: apakah subject dengan nama sama sudah ada?
	dup, err := u.matpelRepo.ExistsByNameAndFeatureID(name, featureID)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, errors.New("mata pelajaran dengan nama tersebut sudah ada pada feature ini")
	}

	if name == "" {
		return nil, errors.New("nama mata pelajaran tidak boleh kosong")
	}

	active := true
	if isActive != nil {
		active = *isActive
	}

	subject, err := u.matpelRepo.Create(featureID, name, deskripsi, active)
	if err != nil {
		return nil, err
	}

	return subject, nil
}

func (u *matpelUsecase) Update(id uint64, featureID uint64, name string, deskripsi *string, isActive bool) (*repository.Matpel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama mata pelajaran wajib diisi")
	}

	exists, err := u.featureRepo.ExistsByID(featureID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("feature_id tidak ditemukan atau tidak aktif")
	}

	dup, err := u.matpelRepo.ExistsByNameAndFeatureIDExceptID(id, featureID, name)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, errors.New("nama mata pelajaran sudah ada pada feature ini")
	}

	updated, err := u.matpelRepo.Update(id, featureID, name, deskripsi, isActive)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (u *matpelUsecase) Delete(id uint64, role string) error {
	if role != "admin" {
		return errors.New("akses ditolak, hanya admin yang dapat menghapus mata pelajaran")
	}

	return u.matpelRepo.Delete(id)
}

func (u *matpelUsecase) GetDetail(id uint64) (*repository.Matpel, error) {
	updated, err := u.matpelRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
