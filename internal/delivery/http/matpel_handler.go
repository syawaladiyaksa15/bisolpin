package http

import (
	"fmt"
	"main-service/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type MatpelHandler struct {
	usecase usecase.MatpelUsecase
}

func NewMatpelHandler(uc usecase.MatpelUsecase) *MatpelHandler {
	return &MatpelHandler{usecase: uc}
}

func (h *MatpelHandler) RegisterRoutes(api fiber.Router) {
	subjects := api.Group("/matpels")
	subjects.Post("/", h.Create)
	subjects.Put("/:id", h.Update)
	subjects.Get("/:feature_id", h.GetByFeatureID)
	subjects.Delete("/:id", h.Delete)
	subjects.Get("/show/:id", h.GetDetail)
}

func (h *MatpelHandler) GetByFeatureID(c *fiber.Ctx) error {
	idParam := c.Params("feature_id")
	featureID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "invalid feature_id parameter",
		})
	}

	subjects, err := h.usecase.GetMatpelByFeature(featureID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status_code": fiber.StatusInternalServerError,
			"status":      "error",
			"message":     "failed to get matpels",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status_code": fiber.StatusOK,
		"status":      "success",
		"data":        subjects,
	})
}

func (h *MatpelHandler) Create(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role == nil || role.(string) != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "forbidden: hanya admin yang dapat menambahkan mata pelajaran",
		})
	}

	type request struct {
		FeatureID uint64  `json:"feature_id"`
		Name      string  `json:"name"`
		Deskripsi *string `json:"deskripsi"`
		IsActive  *bool   `json:"is_active"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "invalid request body",
		})
	}

	// Manual Validation + Trim
	req.Name = strings.TrimSpace(req.Name)
	if req.FeatureID == 0 || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "feature_id dan name wajib diisi",
		})
	}

	if len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "nama minimal 3 karakter",
		})
	}

	subject, err := h.usecase.Create(req.FeatureID, req.Name, req.Deskripsi, req.IsActive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status_code": fiber.StatusInternalServerError,
			"status":      "error",
			"message":     err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status_code": fiber.StatusCreated,
		"status":      "success",
		"data":        subject,
	})
}

func (h *MatpelHandler) Update(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "akses ditolak, hanya admin yang dapat mengedit mata pelajaran",
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "id tidak valid",
		})
	}

	var req struct {
		FeatureID uint64  `json:"feature_id"`
		Name      string  `json:"name"`
		Deskripsi *string `json:"deskripsi"`
		IsActive  bool    `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "input tidak valid",
		})
	}

	// Validasi dasar
	if req.FeatureID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "feature_id wajib diisi",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "nama mata pelajaran wajib diisi",
		})
	}

	matpel, err := h.usecase.Update(id, req.FeatureID, req.Name, req.Deskripsi, req.IsActive)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status_code": fiber.StatusOK,
		"status":      "success",
		"message":     "mata pelajaran berhasil diperbarui",
		"data":        matpel,
	})
}

func (h *MatpelHandler) Delete(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "akses ditolak, hanya admin yang dapat menghapus mata pelajaran",
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "id tidak valid",
		})
	}

	if err := h.usecase.Delete(id, role); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status_code": fiber.StatusOK,
		"status":      "success",
		"message":     "mata pelajaran berhasil dihapus",
	})
}

func (h *MatpelHandler) GetDetail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "id tidak valid",
		})
	}

	matpel, err := h.usecase.GetDetail(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status_code": fiber.StatusOK,
		"status":      "success",
		"message":     fmt.Sprintf("data detail dari id %d", id),
		"data":        matpel,
	})
}
