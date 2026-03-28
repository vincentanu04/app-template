package handler

import (
	"context"

	oapi "myapp/generated/server"
	"myapp/internal/middleware"
	"myapp/internal/repository"
)

func (h *Handler) GetAuthMe(ctx context.Context, _ oapi.GetAuthMeRequestObject) (oapi.GetAuthMeResponseObject, error) {
	userID := middleware.UserIDFromContext(ctx)

	user, err := repository.GetUserByID(ctx, h.deps.DB, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return oapi.GetAuthMe401JSONResponse{Message: "unauthorized"}, nil
	}

	return oapi.GetAuthMe200JSONResponse{Id: user.ID, Email: user.Email}, nil
}
