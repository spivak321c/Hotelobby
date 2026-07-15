package handlers

import (
	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(as *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) writeAuthResponse(c *fiber.Ctx, status int, result *services.AuthResult) error {
	if result.RefreshToken != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    result.RefreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "strict",
			Path:     "/api/auth",
			MaxAge:   7 * 24 * 60 * 60,
		})
	}

	return middleware.Respond(c, status, middleware.OK(fiber.Map{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":         result.User,
	}))
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.authService.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		code := "registration_failed"
		status := fiber.StatusInternalServerError
		if err == services.ErrEmailTaken {
			code = "email_taken"
			status = fiber.StatusConflict
		}
		return middleware.Respond(c, status, middleware.Fail(code, err.Error()))
	}

	return h.writeAuthResponse(c, fiber.StatusCreated, result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_credentials", err.Error()))
	}

	return h.writeAuthResponse(c, fiber.StatusOK, result)
}

func (h *AuthHandler) AdminLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.authService.AdminLogin(c.Context(), req.Email, req.Password)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_credentials", err.Error()))
	}

	return h.writeAuthResponse(c, fiber.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}
	if refreshToken == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_token", "refresh token required"))
	}

	result, err := h.authService.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_token", err.Error()))
	}

	return h.writeAuthResponse(c, fiber.StatusOK, result)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "strict",
		Path:     "/api/auth",
		MaxAge:   -1,
	})
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{"message": "logged out"}))
}
