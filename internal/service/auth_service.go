package service

import (
	"context"
	"errors"

	"time"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/PratikkJadhav/Finance-API/internal/repository"
	"github.com/PratikkJadhav/Finance-API/internal/validator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, error) {
	if err := validator.Validate(input); err != nil {
		return nil, err
	}
	// check if email already exists
	existing, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// default role to viewer
	role := model.RoleViewers
	if input.Role != "" {
		role = model.Role(input.Role)
	}

	user := &model.User{
		Email:    input.Email,
		Password: string(hashed),
		Name:     input.Name,
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// --- Login ---

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResponse, error) {
	// fetch user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// check if account is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// generate JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// --- JWT ---

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateJWT(user *model.User) (string, error) {
	claims := Claims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
