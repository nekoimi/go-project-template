package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	userdomain "github.com/nekoimi/go-project-template/internal/domain/user"
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/repository"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
}

type service struct {
	userRepo  repository.UserRepository
	db        *gorm.DB
	jwtSecret string
	jwtExpire time.Duration
	events    *framework.EventBus
}

func NewService(userRepo repository.UserRepository, db *gorm.DB, jwtSecret string, jwtExpire time.Duration, events *framework.EventBus) Service {
	return &service{
		userRepo:  userRepo,
		db:        db,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
		events:    events,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &userdomain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.userRepo.WithTx(tx)

		if _, err := txRepo.FindByEmail(ctx, req.Email); err == nil {
			return errcode.New(errcode.ErrEmailExists)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if _, err := txRepo.FindByUsername(ctx, req.Username); err == nil {
			return errcode.New(errcode.ErrUsernameExists)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return txRepo.Create(ctx, user)
	})
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	userID := idgen.FormatID(user.ID)
	s.publishUserRegistered(ctx, userID, user.Username, user.Email)

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:        userID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.ErrInvalidCreds)
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errcode.New(errcode.ErrInvalidCreds)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:        idgen.FormatID(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *service) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": idgen.FormatID(userID),
		"exp": time.Now().Add(s.jwtExpire).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) publishUserRegistered(ctx context.Context, userID, username, email string) {
	if s.events == nil {
		return
	}
	_ = s.events.Publish(ctx, framework.Event{
		Topic: EventUserRegistered,
		Payload: UserRegisteredEvent{
			UserID:   userID,
			Username: username,
			Email:    email,
		},
	})
}
