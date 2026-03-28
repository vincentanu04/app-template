package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"myapp/generated/db/myapp_db/public/model"
	"myapp/internal/deps"
	"myapp/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// Login validates credentials and returns the user + a signed JWT.
func Login(ctx context.Context, d deps.Deps, email, password string) (*model.UserTbl, string, error) {
	user, err := repository.GetUserByEmail(ctx, d.DB, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
