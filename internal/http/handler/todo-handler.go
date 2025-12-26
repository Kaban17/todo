package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"todo/internal/domain"
	"todo/internal/usecase"
)

// TodoHandler обрабатывает HTTP запросы для задач
type TodoHandler struct {
	useCase *usecase.TodoUseCase
}

// NewTodoHandler создает новый обработчик
func NewTodoHandler(uc *usecase.TodoUseCase) *TodoHandler {
	return &TodoHandler{
		useCase: uc,
	}
}

// HandleTodos обрабатывает /todos эндпоинт
func (h *TodoHandler) HandleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateTodo(w, r)
	case http.MethodGet:
		h.GetAllTodos(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleTodoByID обрабатывает /todos/{id} эндпоинт
func (h *TodoHandler) HandleTodoByID(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL
	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetTodoByID(w, r, id)
	case http.MethodPut:
		h.UpdateTodo(w, r, id)
	case http.MethodDelete:
		h.DeleteTodo(w, r, id)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// CreateTodo создает новую задачу (POST /todos)
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo domain.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdTodo, err := h.useCase.CreateTodo(r.Context(), &todo)
	if err != nil {
		if errors.Is(err, domain.ErrTodoAlreadyExists) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, createdTodo)
}

// GetAllTodos возвращает все задачи (GET /todos)
func (h *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.useCase.GetAllTodos(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todos")
		return
	}

	respondWithJSON(w, http.StatusOK, todos)
}

// GetTodoByID возвращает задачу по ID (GET /todos/{id})
func (h *TodoHandler) GetTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	todo, err := h.useCase.GetTodoByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todo")
		return
	}

	respondWithJSON(w, http.StatusOK, todo)
}

// UpdateTodo обновляет задачу (PUT /todos/{id})
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var todo domain.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedTodo, err := h.useCase.UpdateTodo(r.Context(), id, &todo)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedTodo)
}

// DeleteTodo удаляет задачу (DELETE /todos/{id})
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	err := h.useCase.DeleteTodo(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete todo")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Вспомогательные функции

func extractIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, errors.New("invalid path")
	}

	return strconv.Atoi(parts[1])
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
