package http

import (
	"fmt"
	"main-service/internal/domain"
	"main-service/internal/repository"
	"main-service/internal/usecase"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BimbelHandler struct {
	Usecase  usecase.BimbelUsecase
	UserRepo repository.UserRepository
}

func NewBimbelHandler(u usecase.BimbelUsecase, ur repository.UserRepository) *BimbelHandler {
	return &BimbelHandler{Usecase: u, UserRepo: ur}
}

// ✅ Daftar semua route handler
func (h *BimbelHandler) RegisterRoutes(api fiber.Router) {
	bimbels := api.Group("/bimbels")
	bimbels.Post("/", h.Create)
	bimbels.Put("/:id", h.Update)
	bimbels.Delete("/:id", h.Delete)
	bimbels.Get("/show/:id", h.GetDetail)
}

// ✅ Helper standardized response
func jsonError(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{
		"status_code": code,
		"status":      "error",
		"message":     msg,
	})
}

func jsonSuccess(c *fiber.Ctx, code int, msg string, data any) error {
	return c.Status(code).JSON(fiber.Map{
		"status_code": code,
		"status":      "success",
		"message":     msg,
		"data":        data,
	})
}

// ✅ CREATE BIMBEL
func (h *BimbelHandler) Create(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(uint64)

	// Ambil data dari form
	name := strings.TrimSpace(c.FormValue("name"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	harga, _ := strconv.ParseFloat(c.FormValue("harga"), 64)
	featureID, _ := strconv.ParseUint(c.FormValue("feature_id"), 10, 64)
	subjectID, _ := strconv.ParseUint(c.FormValue("subject_id"), 10, 64)
	limitPeserta, _ := strconv.Atoi(c.FormValue("limit_peserta"))
	tutorIDForm := c.FormValue("tutor_id")

	// Upload thumbnail
	thumbnailPath, err := saveThumbnail(c)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	// Validasi field wajib
	if name == "" || deskripsi == "" || harga <= 0 || subjectID == 0 || featureID == 0 {
		os.Remove(thumbnailPath)
		return jsonError(c, fiber.StatusBadRequest, "name, deskripsi, harga, feature_id, dan subject_id wajib diisi")
	}

	// Tentukan tutor_id
	var tutorID uint64
	if role == "tutor" {
		user, err := h.UserRepo.FindTutorIDByUserID(userID)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, "gagal mengambil data user")
		}
		if user.TutorID == nil {
			return jsonError(c, fiber.StatusBadRequest, "user belum memiliki tutor_id")
		}
		tutorID = *user.TutorID
	} else if role == "admin" {
		if tutorIDForm == "" {
			os.Remove(thumbnailPath)
			return jsonError(c, fiber.StatusBadRequest, "tutor_id wajib diisi oleh admin")
		}
		tid, err := strconv.ParseUint(tutorIDForm, 10, 64)
		if err != nil {
			os.Remove(thumbnailPath)
			return jsonError(c, fiber.StatusBadRequest, "tutor_id tidak valid")
		}
		tutorID = tid
	} else {
		os.Remove(thumbnailPath)
		return jsonError(c, fiber.StatusForbidden, "role tidak memiliki akses untuk membuat bimbel")
	}

	// Cek nama duplikat
	exists, err := h.Usecase.IsDuplicateName(name, tutorID)
	if err != nil {
		os.Remove(thumbnailPath)
		return jsonError(c, fiber.StatusInternalServerError, err.Error())
	}
	if exists {
		os.Remove(thumbnailPath)
		return jsonError(c, fiber.StatusConflict, "nama bimbel sudah digunakan")
	}

	// Simpan ke database
	bimbel := &domain.Bimbel{
		TutorID:      tutorID,
		FeatureID:    featureID,
		SubjectID:    subjectID,
		Name:         name,
		Deskripsi:    deskripsi,
		Thumbnail:    thumbnailPath,
		Harga:        harga,
		IsActive:     true,
		LimitPeserta: limitPeserta,
	}

	if err := h.Usecase.Create(role, tutorID, bimbel); err != nil {
		os.Remove(thumbnailPath)
		return jsonError(c, fiber.StatusInternalServerError, err.Error())
	}

	return jsonSuccess(c, fiber.StatusCreated, "Bimbel berhasil dibuat", bimbel)
}

// ✅ SAVE THUMBNAIL
func saveThumbnail(c *fiber.Ctx) (string, error) {
	file, err := c.FormFile("thumbnail")
	if err != nil {
		return "", fmt.Errorf("thumbnail wajib diupload")
	}

	// Validasi ekstensi file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("format thumbnail harus jpg, jpeg, atau png")
	}

	// Tentukan direktori penyimpanan absolut
	wd, _ := os.Getwd() // direktori project
	uploadDir := filepath.Join(wd, "uploads", "thumbnails")

	// Buat folder jika belum ada
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("gagal membuat folder upload: %v", err)
	}

	// Buat nama file unik
	filename := fmt.Sprintf("bimbel_%d%s", time.Now().UnixNano(), ext)
	fullPath := filepath.Join(uploadDir, filename)

	// Simpan file ke disk
	if err := c.SaveFile(file, fullPath); err != nil {
		return "", fmt.Errorf("gagal menyimpan file thumbnail: %v", err)
	}

	// Buat URL publik
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
	publicURL := fmt.Sprintf("%s/uploads/thumbnails/%s", baseURL, filename)

	return publicURL, nil
}

// ✅ UPDATE BIMBEL
func (h *BimbelHandler) Update(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userTutorID := c.Locals("tutor_id").(uint64)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	existing, err := h.Usecase.FindByID(role, userTutorID, id)
	if err != nil {
		return jsonError(c, fiber.StatusNotFound, err.Error())
	}

	featureID, _ := strconv.ParseUint(c.FormValue("feature_id"), 10, 64)
	subjectID, _ := strconv.ParseUint(c.FormValue("subject_id"), 10, 64)
	name := c.FormValue("name")
	deskripsi := c.FormValue("deskripsi")
	harga, _ := strconv.ParseFloat(c.FormValue("harga"), 64)

	if name == "" || deskripsi == "" || harga <= 0 || subjectID == 0 {
		return jsonError(c, fiber.StatusBadRequest, "name, deskripsi, subject_id, dan harga wajib diisi")
	}

	thumbnail := existing.Thumbnail
	file, _ := c.FormFile("thumbnail")

	if file != nil {
		// Upload thumbnail baru
		newThumb, err := saveThumbnail(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err.Error())
		}

		// Hapus file lama (jika ada)
		if existing.Thumbnail != "" {
			// Ambil nama file dari URL lama
			parts := strings.Split(existing.Thumbnail, "/uploads/")
			if len(parts) == 2 {
				localPath := filepath.Join("uploads", parts[1])
				_ = os.Remove(localPath) // hapus file fisik
			}
		}

		// Update ke thumbnail baru
		thumbnail = newThumb
	}

	req := &domain.Bimbel{
		ID:           id,
		FeatureID:    featureID,
		SubjectID:    subjectID,
		Name:         name,
		LimitPeserta: existing.LimitPeserta,
		IsActive:     existing.IsActive,
		Thumbnail:    thumbnail,
		Deskripsi:    deskripsi,
		Harga:        harga,
	}

	if err := h.Usecase.Update(role, userTutorID, req); err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	return jsonSuccess(c, fiber.StatusOK, "Bimbel berhasil diperbarui", req)
}

// ✅ DELETE BIMBEL
func (h *BimbelHandler) Delete(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userTutorID := c.Locals("tutor_id").(uint64)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	if err := h.Usecase.Delete(role, userTutorID, id); err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	return jsonSuccess(c, fiber.StatusOK, "Bimbel berhasil dihapus", nil)
}

// ✅ GET DETAIL
func (h *BimbelHandler) GetDetail(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userTutorID := c.Locals("tutor_id").(uint64)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	data, err := h.Usecase.FindByID(role, userTutorID, id)
	if err != nil {
		return jsonError(c, fiber.StatusNotFound, err.Error())
	}

	return jsonSuccess(c, fiber.StatusOK, "Detail bimbel ditemukan", data)
}
