package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Nasaee/go-todo-backend/internal/user"
	"github.com/Nasaee/go-todo-backend/pkg/utils"
)

type Handler struct {
	userService  user.UserService
	tokenService TokenService
	refreshTTL   time.Duration
	isProd       bool
}

func NewHandler(us user.UserService, ts TokenService, refreshTTL time.Duration, isProd bool) *Handler {
	return &Handler{
		userService:  us,
		tokenService: ts,
		refreshTTL:   refreshTTL,
		isProd:       isProd,
	}
}

// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	// อ่าน body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	// สร้าง user ใหม่
	u, err := h.userService.Register(r.Context(), req.FirstName, req.LastName, req.Email, req.Password)
	if err != nil {
		// ตรงนี้คุณจะ map error ให้สวยกว่านี้ทีหลังก็ได้ เช่น เช็ค ErrEmailTaken
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// gen access / refresh token
	access, refresh, err := h.tokenService.GenerateTokens(r.Context(), u.ID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
		return
	}

	// เก็บ refresh_token ลง HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HttpOnly: true,
		Secure:   h.isProd,                    // dev = false, prod = true (อ่านจาก APP_ENV)
		SameSite: http.SameSiteLaxMode,        // กัน CSRF ได้ในระดับนึง
		Path:     "/auth/refresh",             // จะส่ง cookie แค่ตอนเรียก /auth/refresh
		MaxAge:   int(h.refreshTTL.Seconds()), // ใช้ค่าเดียวกับ refresh TTL ใน config
	})

	// ส่ง user + access token กลับไป
	resp := map[string]any{
		"user":         u,
		"access_token": access,
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	u, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	access, refresh, err := h.tokenService.GenerateTokens(r.Context(), u.ID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
		return
	}

	// 🎯 ตั้ง refresh_token เป็น HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HttpOnly: true,
		Secure:   h.isProd,                    // dev = false, prod = true
		SameSite: http.SameSiteLaxMode,        // กัน CSRF ได้พอสมควร
		Path:     "/auth/refresh",             // ส่ง cookie แค่ตอนเรียก /auth/refresh
		MaxAge:   int(h.refreshTTL.Seconds()), // ใช้ค่าเดียวกับ refresh token TTL
	})

	resp := map[string]any{
		"user":         u,
		"access_token": access,
		// ไม่ส่ง refresh_token ใน body แล้ว เพราะอยู่ใน cookie แล้ว
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	access, newRefresh, err := h.tokenService.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		switch err {
		case ErrExpiredRefreshToken:
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh_token_expired", "message": "Please login again."})
		case ErrInvalidRefreshToken:
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_refresh_token", "message": "Please login again."})
		default:
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh_token_error", "message": "Please login again."})
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefresh,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/refresh",
		MaxAge:   int(h.refreshTTL.Seconds()),
	})

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"access_token": access,
	})
}
