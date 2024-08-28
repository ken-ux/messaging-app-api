package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init() error {
	// Connect to database.
	var dbpool_err error
	Pool, dbpool_err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if dbpool_err != nil {
		return fmt.Errorf("unable to create connection pool: %v", dbpool_err)
	}
	// defer Pool.Close()

	return nil
}
