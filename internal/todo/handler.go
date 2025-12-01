package todo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Nasaee/go-todo-backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// =============== context + helper =================

// ตรงนี้ให้แก้ให้ match กับ auth middleware ของเพื่อน
// สมมติ middleware set ctx.Value("userID") = int64(userID)
type ctxKey string

const userIDCtxKey ctxKey = "userID"

func userIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(userIDCtxKey)
	if val == nil {
		return 0, errors.New("user not authenticated")
	}

	userID, ok := val.(int64)
	if !ok {
		return 0, errors.New("invalid user id type in context")
	}

	if userID == 0 {
		return 0, errors.New("invalid user id")
	}

	return userID, nil
}

type errorResponse struct {
	Error string `json:"error"`
}

// =============== Handler struct ==================

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/todos
func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	todos, err := h.svc.ListTodos(ctx, userID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "failed to list todos")
		return
	}

	utils.WriteJSON(w, http.StatusOK, todos)
}

// GET /api/todos/today
func (h *Handler) ListTodayTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	todos, err := h.svc.ListTodayTodos(ctx, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to list today todos")
		return
	}

	utils.WriteJSON(w, http.StatusOK, todos)
}

// GET /api/todos/tomorrow
func (h *Handler) ListTomorrow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	todos, err := h.svc.ListTomorrowTodos(ctx, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to list tomorrow todos")
		return
	}

	utils.WriteJSON(w, http.StatusOK, todos)
}

// GET /api/todos/this-week
func (h *Handler) ListThisWeek(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	todos, err := h.svc.ListThisWeekTodos(ctx, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to list this week todos")
		return
	}

	utils.WriteJSON(w, http.StatusOK, todos)
}

// GET /api/todos/{id}
func (h *Handler) GetTodoByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	todo, err := h.svc.GetTodo(ctx, id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, "todo not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to get todo")
		return
	}

	utils.WriteJSON(w, http.StatusOK, todo)
}

// POST /api/todos
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var in CreateTodoInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	todo, err := h.svc.CreateTodo(ctx, userID, in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidDateRange):
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, "failed to create todo")
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, todo)
}

// PUT /api/todos/{id}
func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var in UpdateTodoInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	todo, err := h.svc.UpdateTodo(ctx, id, userID, in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidDateRange):
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, ErrNotFound):
			utils.WriteError(w, http.StatusNotFound, "todo not found")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, "failed to update todo")
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, todo)
}

// DELETE /api/todos/{id}
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromContext(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.svc.DeleteTodo(ctx, id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, "todo not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to delete todo")
		return
	}

	// ถ้าอยากให้ body ว่าง ๆ ใช้ StatusNoContent ก็ได้
	utils.WriteJSON(w, http.StatusNoContent, nil)
}
