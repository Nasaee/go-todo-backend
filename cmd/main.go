package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Nasaee/go-todo-backend/internal/auth"
	"github.com/Nasaee/go-todo-backend/internal/env"
	"github.com/Nasaee/go-todo-backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	env.Init()

	// base context (อนาคตถ้าจะทำ cancel เองก็ทำจากตรงนี้ได้)
	ctx := context.Background()

	cfg := config{
		addr: env.GetString("API_PORT", ":8000"),
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost port=5433 user=postgres password=P@ssw0rd dbname=ecom sslmode=disable"),
		},
	}

	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database connect
	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	logger.Info("database pool connected 🎉")
	defer pool.Close()

	// redis
	rdb := redis.NewClient(&redis.Options{
		Addr: env.GetString("REDIS_ADDR", "localhost:6379"),
	})
	defer rdb.Close()

	// services
	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)

	refreshTTL := 7 * 24 * time.Hour
	accessTTL := 15 * time.Minute

	isProd := env.GetString("APP_ENV", string(env.EnvDevelopment)) == string(env.EnvProduction)

	tokenSvc := auth.NewTokenService(
		env.GetString("JWT_SECRET", "dev-secret"),
		accessTTL,
		refreshTTL,
		rdb,
	)

	api := application{
		config:       cfg,
		db:           pool,
		userService:  userSvc,
		tokenService: tokenSvc,
		refreshTTL:   refreshTTL,
		isProd:       isProd,
	}

	// ใช้ ctx + graceful shutdown
	if err := api.run(ctx, api.mount()); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
