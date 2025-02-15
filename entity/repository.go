package entity

import (
	"context"
	"database/sql"
)

// TransactionFunc defines a function that executes inside a transaction.
type TransactionFunc func(ctx context.Context, tx *sql.Tx) error
