package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"subscriptions-service/internal"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	up := flag.Bool("up", false, "run up migrations")
	flag.Parse()

	if *up {
		if err := runMigrations(); err != nil {
			log.Fatal(err)
		}
		log.Println("migrations applied successfully")
	}
}

func runMigrations() error {
	cfg, err := internal.NewConfig[internal.AppConfig](".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresConfig.DBAdapter,
		cfg.PostgresConfig.DBUser,
		cfg.PostgresConfig.DBPassword,
		cfg.PostgresConfig.DBHost,
		cfg.PostgresConfig.DBPort,
		cfg.PostgresConfig.DBName,
		cfg.PostgresConfig.DBSSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	files, err := filepath.Glob(filepath.Join("migrations", "*.up.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", file, err)
		}
		log.Printf("applied: %s", file)
	}

	return nil
}
