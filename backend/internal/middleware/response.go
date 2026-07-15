package middleware

import "github.com/gofiber/fiber/v2"

// Response is the standard API envelope for all endpoints.
// PRD §80: { success, data?, error?: { code, message } }
type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo carries a machine-readable code and a human message.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK wraps data in a success envelope.
func OK(data any) Response {
	return Response{Success: true, Data: data}
}

// Created wraps data in a success envelope (used with 201 status).
func Created(data any) Response {
	return Response{Success: true, Data: data}
}

// Fail builds an error envelope.
func Fail(code, message string) Response {
	return Response{Success: false, Error: &ErrorInfo{Code: code, Message: message}}
}

// ErrResp is a shorthand fiber.Map for inline error responses.
func ErrResp(code, message string) fiber.Map {
	return fiber.Map{
		"success": false,
		"error":   fiber.Map{"code": code, "message": message},
	}
}

// Respond writes status + JSON envelope in one call.
func Respond(c *fiber.Ctx, status int, resp Response) error {
	return c.Status(status).JSON(resp)
}
