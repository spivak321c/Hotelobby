package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ParseAndValidate parses the request body into the given struct and runs
// struct tag validation. Returns a 422 response with a machine-readable
// error code and field-level details on failure.
func ParseAndValidate(c *fiber.Ctx, out any) error {
	if err := c.BodyParser(out); err != nil {
		return Respond(c, fiber.StatusUnprocessableEntity, Fail("invalid_request", "invalid request body"))
	}

	if err := validate.Struct(out); err != nil {
		code := "validation_failed"
		if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
			code = "validation_" + ve[0].Field() + "_" + ve[0].Tag()
		}
		return Respond(c, fiber.StatusUnprocessableEntity, Fail(code, err.Error()))
	}

	return nil
}
