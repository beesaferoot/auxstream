package db

import (
	"context"

	"github.com/google/uuid"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
}

type userRepo struct {
	Db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{
		Db: db,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user *User) (*User, error) {
	if err := validator.Validate(user); err != nil {
		return nil, err
	}

	res := r.Db.WithContext(ctx).Create(user)

	return user, res.Error
}

func (r *userRepo) GetUserById(ctx context.Context, id uuid.UUID) (*User, error) {
	user := &User{}
	res := r.Db.WithContext(ctx).First(user, id)
	return user, res.Error
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	res := r.Db.WithContext(ctx).Where("username = ?", username).First(user)
	return user, res.Error
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	res := r.Db.WithContext(ctx).Where("email = ?", email).First(user)
	return user, res.Error
}

func (r *userRepo) GetUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	user := &User{}
	res := r.Db.WithContext(ctx).Where("google_id = ?", googleID).First(user)
	return user, res.Error
}

func (r *userRepo) UpdateUser(ctx context.Context, user *User) (*User, error) {
	res := r.Db.WithContext(ctx).Save(user)
	return user, res.Error
}
