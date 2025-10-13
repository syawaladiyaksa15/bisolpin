package http

import (
	"fmt"
	"main-service/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type FeatureHandler struct {
	usecase usecase.FeatureUsecase
}

func NewFeatureHandler(uc usecase.FeatureUsecase) *FeatureHandler {
	return &FeatureHandler{usecase: uc}
}

func (h *FeatureHandler) RegisterRoutes(api fiber.Router) {
	features := api.Group("/features")
	features.Get("/", h.GetFeatures)
	features.Post("/", h.Create)
	features.Put("/:id", h.Update)
	features.Delete("/:id", h.Delete)
	features.Get("/show/:id", h.GetDetail)
}

func (h *FeatureHandler) GetFeatures(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status_code": fiber.StatusUnauthorized,
			"status":      "error",
			"message":     "unauthorized",
		})
	}

	features, err := h.usecase.GetFeaturesByRole(role.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status_code": fiber.StatusInternalServerError,
			"status":      "error",
			"message":     err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status_code": fiber.StatusOK,
		"status":      "success",
		"message":     "daftar fitur untuk role " + role.(string),
		"data":        features,
	})
}

func (h *FeatureHandler) Create(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role == nil || role.(string) != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "forbidden: hanya admin yang dapat menambahkan mata pelajaran",
		})
	}

	type request struct {
		Name     string `json:"name"`
		Roles    string `json:"roles"`
		IsActive *bool  `json:"is_active"`
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
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "name wajib diisi",
		})
	}

	if len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "nama minimal 3 karakter",
		})
	}

	req.Roles = strings.TrimSpace(req.Roles)
	if req.Roles == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "roles wajib diisi",
		})
	}

	if len(req.Roles) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "roles minimal 5 karakter",
		})
	}

	feature, err := h.usecase.Create(req.Name, req.Roles, req.IsActive)
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
		"data":        feature,
	})
}

func (h *FeatureHandler) Update(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "akses ditolak, hanya admin yang dapat mengedit fitur",
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
		Name     string `json:"name"`
		Roles    string `json:"roles"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "input tidak valid",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "nama fitur wajib diisi",
		})
	}

	if strings.TrimSpace(req.Roles) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "roles wajib diisi",
		})
	}

	feature, err := h.usecase.Update(id, req.Name, req.Roles, req.IsActive)
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
		"message":     "fitur berhasil diperbarui",
		"data":        feature,
	})
}

func (h *FeatureHandler) Delete(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status_code": fiber.StatusForbidden,
			"status":      "error",
			"message":     "akses ditolak, hanya admin yang dapat menghapus fitur",
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
		"message":     "fitur berhasil dihapus",
	})
}

func (h *FeatureHandler) GetDetail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"status":      "error",
			"message":     "id tidak valid",
		})
	}

	feature, err := h.usecase.GetDetail(id)
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
		"data":        feature,
	})
}
