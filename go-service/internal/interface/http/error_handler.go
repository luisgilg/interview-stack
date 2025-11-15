package http

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

func handleError(ctx *fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return ctx.Status(fiberErr.Code).JSON(dto.ErrorResponse{Error: fiberErr.Message})
	}
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case domain.ErrorCodeValidation:
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: domainErr.Message})
		case domain.ErrorCodeNotFound:
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: domainErr.Message})
		default:
			return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: domainErr.Message})
		}
	}
	return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "internal server error"})
}
