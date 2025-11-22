package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nasaee/go-todo-backend/internal/auth"
	"github.com/Nasaee/go-todo-backend/internal/user"
	"github.com/Nasaee/go-todo-backend/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

type dbConfig struct {
	dsn string
}

type config struct {
	addr string
	db   dbConfig
}

type application struct {
	config       config
	db           *pgx.Conn
	userService  user.UserService
	tokenService auth.TokenService
	refreshTTL   time.Duration
	isProd       bool
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID) // important for rate limiting
	r.Use(middleware.RealIP)    // important for rate limiting and analytics and tracing
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // recover from panics or crashes

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	authHandler := auth.NewHandler(app.userService, app.tokenService, app.refreshTTL, app.isProd)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	// ส่วนนี้คือ protected routes (ต้องมี access token)
	r.Route("/", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// ใช้ AuthMiddleware ครอบทั้ง group
			r.Use(auth.AuthMiddleware(app.tokenService))

			// ตัวอย่าง route: GET /api/me
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				userID, ok := auth.UserIDFromContext(r.Context())
				if !ok {
					http.Error(w, "user not found in context", http.StatusUnauthorized)
					return
				}

				// ตรงนี้อนาคตคุณค่อยไปดึง user จาก DB
				utils.WriteJSON(w, http.StatusOK, map[string]any{
					"user_id": userID,
				})
			})

			// TODO: อนาคต /api/todos ก็มาอยู่ใน group นี้
		})
	})

	return r
}

/*
	Graceful shutdown:
	- ปกติ ListenAndServe() = บล็อก รันไปเรื่อย ๆ จนกว่าจะ error
	- Graceful ทำให้เราคุมมันเอง:
		1. รัน ListenAndServe() ใน goroutine
		2. เรียก srv.Shutdown(ctx)
		3. เมื่อมี signal:
			- เลิกรับ request ใหม่
			- ปล่อยให้ request ที่กำลังรันอยู่จบ (ในเวลาที่เหลือใน ctx)
			- ปิด listener ต่าง ๆ ให้เรียบร้อย
	ใช้กับ production (Docker, k8s, VM) ดีมาก เพราะตอน rollout version ใหม่ หรือ scale down จะไม่ตัด request กลางคัน
*/

func (app *application) run(ctx context.Context, h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	// channel ไว้รับ error จาก ListenAndServe
	errCh := make(chan error, 1)

	// รัน server ใน goroutine
	go func() {
		slog.Info("starting server", "addr", app.config.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// รอ signal จาก OS หรือ ctx ถูก cancel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-ctx.Done():
		slog.Info("context cancelled, shutting down server...")
	case sig := <-quit:
		slog.Info("received shutdown signal", "signal", sig.String())
	case err := <-errCh:
		// ถ้า server ตายเองก่อน (เช่น listen พัง) → คืน error กลับไปเลย
		if err != nil {
			return err
		}
		// ถ้า err == nil = server ปิดเองอย่างเรียบร้อย
		return nil
	}

	// ถึงตรงนี้คือเราเลือกจะ shutdown เอง (จาก ctx หรือ signal)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		return err
	}

	slog.Info("server exited gracefully")
	return nil
}
