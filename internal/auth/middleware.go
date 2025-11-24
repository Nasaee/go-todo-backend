package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Nasaee/go-todo-backend/pkg/utils"
)

// ---- context key สำหรับเก็บ userID ----

type ctxKey string

const ctxKeyUserID ctxKey = "userID"

func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxKeyUserID)
	if v == nil {
		return 0, false
	}

	id, ok := v.(int64)
	return id, ok
}

// ---- AuthMiddleware สำหรับตรวจ access token ----

func AuthMiddleware(ts TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) ดึง Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimSpace(authHeader[len(prefix):])
			if tokenStr == "" {
				http.Error(w, "empty bearer token", http.StatusUnauthorized)
				return
			}

			// 2) ตรวจ access token
			claims, err := ts.ParseAccessToken(r.Context(), tokenStr)
			if err != nil {
				switch err {
				case ErrExpiredAccessToken:
					utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
						"message": "access token expired",
					})
				case ErrInvalidAccessToken:
					utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
						"message": "invalid access token",
					})
				default:
					utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
						"message": "access token error",
					})
				}
				return
			}

			// 3) ยัด userID ลง context
			ctx := ContextWithUserID(r.Context(), claims.UserID)

			// 4) ส่งต่อให้ handler ถัดไป
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
