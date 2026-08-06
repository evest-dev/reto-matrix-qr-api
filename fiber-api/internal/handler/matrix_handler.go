package handler

import (
	"errors"

	"fiber-api/internal/domain"
	"fiber-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// MatrixHandler expone el caso de uso a través de HTTP.
type MatrixHandler struct {
	processor *service.MatrixProcessor
}

// NewMatrixHandler arma el handler con su procesador.
func NewMatrixHandler(p *service.MatrixProcessor) *MatrixHandler {
	return &MatrixHandler{processor: p}
}

// Register agrega las rutas de /matrix.
func (h *MatrixHandler) Register(router fiber.Router) {
	group := router.Group("/matrix")
	group.Post("/process", h.Process)
	group.Post("/rotate", h.Rotate)
}

// Process atiende POST /matrix/process.
func (h *MatrixHandler) Process(c *fiber.Ctx) error {
	var req ProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "el cuerpo de la petición no es un JSON válido")
	}

	result, err := h.processor.Process(c.UserContext(), req.Matrix)
	if err != nil {
		return mapDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(NewProcessResponse(result))
}

// Rotate atiende POST /matrix/rotate: solo rota, no factoriza.
func (h *MatrixHandler) Rotate(c *fiber.Ctx) error {
	var req ProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "el cuerpo de la petición no es un JSON válido")
	}

	if err := req.Matrix.Validate(); err != nil {
		return mapDomainError(c, err)
	}

	rotated := service.Rotate90Clockwise(req.Matrix)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"original": req.Matrix,
		"rotated":  rotated,
	})
}

// mapDomainError es el único punto donde el dominio se traduce a HTTP.
func mapDomainError(c *fiber.Ctx, err error) error {
	var inconsistent *domain.InconsistentRowError

	switch {
	case errors.Is(err, domain.ErrEmptyMatrix),
		errors.Is(err, domain.ErrNotFactorizable),
		errors.As(err, &inconsistent):
		return respondError(c, fiber.StatusBadRequest, err.Error())
	default:
		return respondError(c, fiber.StatusInternalServerError, "error procesando la matriz")
	}
}

func respondError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
