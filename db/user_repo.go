package db

import (
	"context"
	"strconv"

	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	GetUserById(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

type userRepo struct {
	Db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{
		Db: db,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	user := &User{Username: username, PasswordHash: passwordHash}

	if err := validator.Validate(user); err != nil {
		return nil, err
	}

	res := r.Db.WithContext(ctx).Create(user)

	return user, res.Error
}

func (r *userRepo) GetUserById(ctx context.Context, id string) (*User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}

	user := &User{}
	res := r.Db.WithContext(ctx).First(user, userID)
	return user, res.Error
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	res := r.Db.WithContext(ctx).Where("username = ?", username).First(user)
	return user, res.Error
}
