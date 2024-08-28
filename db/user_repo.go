package db

import (
	"context"

	"gopkg.in/validator.v2"
)

type UserRepo interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	GetUserById(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

type userRepo struct {
	Db DbConn
}

func NewUserRepo(db DbConn) UserRepo {
	return &userRepo{
		Db: db,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	user := &User{Username: username, PasswordHash: passwordHash}

	if err := validator.Validate(user); err != nil {
		return nil, err
	}
	
	stmt := `INSERT INTO auxstream.users (username, password_hash)
			 VALUES ($1, $2)
			 RETURNING id, created_at
			 `
	row := r.Db.QueryRow(ctx, stmt, user.Username, user.PasswordHash)
	err := row.Scan(&user.Id, &user.CreatedAt)
	return user, err
}

func (r *userRepo) GetUserById(ctx context.Context, id string) (*User, error) {
	user := &User{}
	stmt := `SELECT id, username, password_hash, created_at
 			 FROM auxstream.users
 			 WHERE id = $1`
	row := r.Db.QueryRow(ctx, stmt, id)

	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt)

	return user, err
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	stmt := `SELECT id, username, password_hash, created_at
 			 FROM auxstream.users
 			 WHERE username = $1`
	row := r.Db.QueryRow(ctx, stmt, username)

	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt)

	return user, err
}
