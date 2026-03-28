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

type postAuthRegisterCookieResponse struct {
	user  oapi.AuthUser
	token string
}

func (r postAuthRegisterCookieResponse) VisitPostAuthRegisterResponse(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    r.token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	return json.NewEncoder(w).Encode(r.user)
}

func (h *Handler) PostAuthRegister(ctx context.Context, request oapi.PostAuthRegisterRequestObject) (oapi.PostAuthRegisterResponseObject, error) {
	if request.Body.Email == "" || request.Body.Password == "" {
		return oapi.PostAuthRegister400JSONResponse{Message: "email and password are required"}, nil
	}

	user, token, err := auth.Register(ctx, h.deps, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			return oapi.PostAuthRegister409JSONResponse{Message: "email already registered"}, nil
		}
		return nil, err
	}

	return postAuthRegisterCookieResponse{
		user:  oapi.AuthUser{Id: user.ID, Email: user.Email},
		token: token,
	}, nil
}
