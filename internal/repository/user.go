package repository

import (
	"context"

	"gorm.io/gorm"

	userdomain "github.com/nekoimi/go-project-template/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, user *userdomain.User) error
	FindByID(ctx context.Context, id int64) (*userdomain.User, error)
	FindByEmail(ctx context.Context, email string) (*userdomain.User, error)
	FindByUsername(ctx context.Context, username string) (*userdomain.User, error)
	WithTx(tx *gorm.DB) UserRepository
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) WithTx(tx *gorm.DB) UserRepository {
	return &userRepo{db: tx}
}

func (r *userRepo) Create(ctx context.Context, user *userdomain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*userdomain.User, error) {
	var user userdomain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	var user userdomain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*userdomain.User, error) {
	var user userdomain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
