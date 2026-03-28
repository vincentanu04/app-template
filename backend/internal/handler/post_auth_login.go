package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	oapi "myapp/generated/server"
	"myapp/internal/app_service/auth"
	"myapp/internal/middleware"
)

type postAuthLoginCookieResponse struct {
	user  oapi.AuthUser
	token string
}

func (r postAuthLoginCookieResponse) VisitPostAuthLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    r.token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	return json.NewEncoder(w).Encode(r.user)
}

func (h *Handler) PostAuthLogin(ctx context.Context, request oapi.PostAuthLoginRequestObject) (oapi.PostAuthLoginResponseObject, error) {
	user, token, err := auth.Login(ctx, h.deps, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return oapi.PostAuthLogin401JSONResponse{Message: "invalid credentials"}, nil
		}
		return nil, err
	}

	return postAuthLoginCookieResponse{
		user:  oapi.AuthUser{Id: user.ID, Email: user.Email},
		token: token,
	}, nil
}
