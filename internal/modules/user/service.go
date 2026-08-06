package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/repository"
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*UserResponse, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*UserResponse, error) {
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

	return &UserResponse{
		ID:        idgen.FormatID(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
