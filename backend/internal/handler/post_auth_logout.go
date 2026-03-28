package handler

import (
	"context"
	"net/http"

	oapi "myapp/generated/server"
	"myapp/internal/middleware"
)

type logoutClearCookieResponse struct{}

func (r logoutClearCookieResponse) VisitPostAuthLogoutResponse(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})
	w.WriteHeader(204)
	return nil
}

func (h *Handler) PostAuthLogout(ctx context.Context, _ oapi.PostAuthLogoutRequestObject) (oapi.PostAuthLogoutResponseObject, error) {
	return logoutClearCookieResponse{}, nil
}
