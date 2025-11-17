package database

import (
	"context"
	"embed"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		content, err := migrationFiles.ReadFile("migrations/" + f.Name())
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return err
		}
	}
	return nil
}
