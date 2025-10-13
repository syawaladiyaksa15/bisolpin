package http

import (
	"main-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	usecase usecase.UserUsecase
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

func (h *UserHandler) RegisterRoutes(api fiber.Router) {
	api.Post("/login", h.Login)
	api.Post("/register", h.Register)
}

// standardized response helper
func response(c *fiber.Ctx, statusCode int, status string, message string, data interface{}) error {
	res := fiber.Map{
		"status_code": statusCode,
		"status":      status,
		"message":     message,
	}

	if data != nil {
		res["data"] = data
	}

	return c.Status(statusCode).JSON(res)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response(c, fiber.StatusBadRequest, "error", "invalid request payload", nil)
	}

	result, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		return response(c, fiber.StatusUnauthorized, "error", err.Error(), nil)
	}

	return response(c, fiber.StatusOK, "success", "login berhasil", result)
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response(c, fiber.StatusBadRequest, "error", "invalid request payload", nil)
	}

	result, err := h.usecase.Register(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		return response(c, fiber.StatusBadRequest, "error", err.Error(), nil)
	}

	return response(c, fiber.StatusCreated, "success", "registrasi berhasil", result)
}
