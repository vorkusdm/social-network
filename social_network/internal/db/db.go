package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init() error {
	connString := os.Getenv("DB_URL")
	if connString == "" {
		connString = "postgres://user:password@localhost:5432/social_network"
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	
	fmt.Println("Successfully connected to PostgreSQL")
	return nil
}