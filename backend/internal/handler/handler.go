package handler

import "myapp/internal/deps"

type Handler struct {
	deps deps.Deps
}

func NewHandler(deps deps.Deps) *Handler {
	return &Handler{deps: deps}
}
