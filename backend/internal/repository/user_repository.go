package repository

import (
	"context"

	"github.com/google/uuid"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"myapp/generated/db/myapp_db/public/model"
	"myapp/generated/db/myapp_db/public/table"
)

// GetUserByEmail returns a user matching the given email, or nil if not found.
func GetUserByEmail(ctx context.Context, db qrm.DB, email string) (*model.UserTbl, error) {
	tbl := table.UserTbl

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.Email.EQ(pg.String(email)))

	var rows []model.UserTbl
	if err := stmt.QueryContext(ctx, db, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// GetUserByID returns a user by primary key, or nil if not found.
func GetUserByID(ctx context.Context, db qrm.DB, id uuid.UUID) (*model.UserTbl, error) {
	tbl := table.UserTbl

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.ID.EQ(pg.UUID(id)))

	var rows []model.UserTbl
	if err := stmt.QueryContext(ctx, db, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
