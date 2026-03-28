package handler

import (
	"context"

	oapi "myapp/generated/server"
)

func (h *Handler) GetHealth(ctx context.Context, _ oapi.GetHealthRequestObject) (oapi.GetHealthResponseObject, error) {
	return oapi.GetHealth200JSONResponse{Status: "ok"}, nil
}
