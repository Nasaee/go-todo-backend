package todogroup

import (
	"encoding/json"
	"net/http"

	"github.com/Nasaee/go-todo-backend/internal/auth"
	"github.com/Nasaee/go-todo-backend/pkg/utils"
)

type Handler struct {
	svc TodoGroupService
}

func NewHandler(svc TodoGroupService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// ดึง user id จาก context (มาจาก AuthMiddleware)
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	var input CreateTodoGroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"message": "invalid body",
		})
		return
	}

	g, err := h.svc.Create(r.Context(), userID, input)
	if err != nil {
		if err == ErrEmptyName {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"message": "name is required",
			})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "could not create todo group",
		})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, g)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	groups, err := h.svc.GetAll(r.Context(), userID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "could not fetch todo groups",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, groups)
}
