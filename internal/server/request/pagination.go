package request

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
)

func GetPagination(ctx fiber.Ctx) (int, int, error) {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return 0, 0, fiber.NewError(400, "page must be a number")
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "25"))
	if err != nil {
		return 0, 0, fiber.NewError(400, "limit must be a number")
	}

	if page < 1 {
		return 0, 0, fiber.NewError(400, "page must be at least 1")
	}

	if limit < 1 || limit > 100 {
		return 0, 0, fiber.NewError(400, "limit must be between 1 and 100")
	}

	return page, limit, nil
}

func GetCursor(ctx fiber.Ctx) (*database.Cursor, error) {
	before := ctx.Query("before")
	if before == "" {
		return nil, nil
	}

	cursor, err := database.ParseCursor(before)
	if err != nil {
		return nil, fiber.NewError(400, "Invalid cursor")
	}

	return cursor, nil
}
