package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/model"
	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/repository"
)

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*model.UserResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*model.UserResponse, error) {
	uid, err := idgen.ParseID(userID)
	if err != nil {
		return nil, errcode.New(errcode.Unauthorized)
	}

	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.ErrUserNotFound)
		}
		return nil, err
	}

	return &model.UserResponse{
		ID:        idgen.FormatID(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
