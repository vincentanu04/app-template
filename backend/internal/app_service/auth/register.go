package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"myapp/generated/db/myapp_db/public/model"
	"myapp/internal/deps"
	"myapp/internal/mutation"
	"myapp/internal/repository"
)

var ErrEmailTaken = errors.New("email already registered")

// Register creates a new user account and returns the user + a signed JWT.
func Register(ctx context.Context, d deps.Deps, email, password string) (*model.UserTbl, string, error) {
	existing, err := repository.GetUserByEmail(ctx, d.DB, email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user, err := mutation.InsertUser(ctx, d.DB, email, string(hash))
	if err != nil {
		return nil, "", err
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
