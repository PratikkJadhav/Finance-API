package service

import (
	"context"
	"errors"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/PratikkJadhav/Finance-API/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *UserService) UpdateRole(ctx context.Context, id string, role string) error {
	validRoles := map[string]bool{
		"viewer":  true,
		"analyst": true,
		"admin":   true,
	}
	if !validRoles[role] {
		return errors.New("invalid role, must be viewer, analyst or admin")
	}

	return s.userRepo.UpdateRole(ctx, id, model.Role(role))
}

func (s *UserService) UpdateStatus(ctx context.Context, id string, isActive bool) error {
	return s.userRepo.UpdateStatus(ctx, id, isActive)
}
