package mutation

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/go-jet/jet/v2/qrm"

	"myapp/generated/db/myapp_db/public/model"
	"myapp/generated/db/myapp_db/public/table"
)

// InsertUser creates a new user record.
func InsertUser(ctx context.Context, db qrm.DB, email, passwordHash string) (*model.UserTbl, error) {
	tbl := table.UserTbl

	record := model.UserTbl{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	stmt := tbl.INSERT(tbl.MutableColumns).
		MODEL(record).
		RETURNING(tbl.AllColumns)

	var dest []model.UserTbl
	if err := stmt.QueryContext(ctx, db, &dest); err != nil {
		return nil, err
	}
	if len(dest) != 1 {
		return nil, nil
	}
	return &dest[0], nil
}
