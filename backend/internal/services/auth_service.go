package services

import (
	"context"
	"errors"
	"time"

	"hotel_lobby/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type AuthService struct {
	customerRepo CustomerRepository
	adminRepo    AdminRepository
	jwtSecret    []byte
}

type CustomerRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	FindByEmail(ctx context.Context, email string) (*models.Customer, error)
	Create(ctx context.Context, c *models.Customer) error
}

type AdminRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	FindByEmail(ctx context.Context, email string) (*models.Admin, error)
	Create(ctx context.Context, a *models.Admin) error
}

func NewAuthService(cr CustomerRepository, ar AdminRepository, jwtSecret string) *AuthService {
	return &AuthService{
		customerRepo: cr,
		adminRepo:    ar,
		jwtSecret:    []byte(jwtSecret),
	}
}

type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*AuthResult, error) {
	existing, _ := s.customerRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	customer := &models.Customer{
		FullName:     name,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	result, err := s.generateTokens(customer.ID.String(), "customer")
	if err != nil {
		return nil, err
	}
	result.User = customer
	return result, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	customer, err := s.customerRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	result, err := s.generateTokens(customer.ID.String(), "customer")
	if err != nil {
		return nil, err
	}
	result.User = customer
	return result, nil
}

func (s *AuthService) AdminLogin(ctx context.Context, email, password string) (*AuthResult, error) {
	admin, err := s.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !admin.IsActive {
		return nil, ErrInvalidCredentials
	}

	result, err := s.generateTokens(admin.ID.String(), admin.Role)
	if err != nil {
		return nil, err
	}
	result.User = admin
	return result, nil
}

func (s *AuthService) generateTokens(userID, role string) (*AuthResult, error) {
	accessToken, err := s.generateJWT(userID, role, AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateJWT(userID, role, RefreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateJWT(userID, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenStr string) (userID, role string, err error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return "", "", err
	}
	return claims.UserID, claims.Role, nil
}

func (s *AuthService) parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Refresh tokens can be for customer or admin — just reissue with same claims
	return s.generateTokens(claims.UserID, claims.Role)
}
