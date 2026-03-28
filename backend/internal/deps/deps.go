package deps

import "database/sql"

// Deps holds shared application dependencies injected through the server.
// Add additional dependencies (e.g. Redis client, email sender) here as needed.
type Deps struct {
	DB *sql.DB
}
