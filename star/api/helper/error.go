package helper

import "github.com/gofiber/fiber/v2"

func NewHttpError(code int, message string) *fiber.Error {
	return &fiber.Error{
		Code: code,
		Message: message,
	}
}
