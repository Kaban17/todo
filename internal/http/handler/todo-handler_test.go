package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"todo/internal/domain"
	"todo/internal/repository"
	"todo/internal/usecase"
)

func setupTestHandler() *TodoHandler {
	repo := repository.NewInMemoryTodoRepository()
	uc := usecase.NewTodoUseCase(repo)
	return NewTodoHandler(uc)
}

func TestTodoHandler_CreateTodo(t *testing.T) {
	handler := setupTestHandler()

	t.Run("успешное создание задачи", func(t *testing.T) {
		todo := domain.Todo{
			Title:       "Test Todo",
			Description: "Test Description",
			Completed:   false,
		}

		body, _ := json.Marshal(todo)
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.HandleTodos(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var created domain.Todo
		json.NewDecoder(rec.Body).Decode(&created)

		if created.ID == 0 {
			t.Error("expected ID to be assigned")
		}

		if created.Title != todo.Title {
			t.Errorf("expected title %s, got %s", todo.Title, created.Title)
		}
	})

	t.Run("создание задачи с пустым заголовком", func(t *testing.T) {
		todo := domain.Todo{
			Title:       "",
			Description: "Test",
		}

		body, _ := json.Marshal(todo)
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.HandleTodos(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("создание задачи с некорректным JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString("invalid json"))
		rec := httptest.NewRecorder()

		handler.HandleTodos(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("создание задачи с дублирующим ID", func(t *testing.T) {
		// Создаем первую задачу
		todo1 := domain.Todo{ID: 999, Title: "First"}
		body1, _ := json.Marshal(todo1)
		req1 := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body1))
		rec1 := httptest.NewRecorder()
		handler.HandleTodos(rec1, req1)

		// Пытаемся создать вторую задачу с тем же ID
		todo2 := domain.Todo{ID: 999, Title: "Second"}
		body2, _ := json.Marshal(todo2)
		req2 := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body2))
		rec2 := httptest.NewRecorder()
		handler.HandleTodos(rec2, req2)

		if rec2.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, rec2.Code)
		}
	})
}

func TestTodoHandler_GetAllTodos(t *testing.T) {
	handler := setupTestHandler()

	// Создаем несколько задач
	for i := 1; i <= 3; i++ {
		todo := domain.Todo{Title: fmt.Sprintf("Todo %d", i)}
		body, _ := json.Marshal(todo)
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.HandleTodos(rec, req)
	}

	// Получаем все задачи
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()

	handler.HandleTodos(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var todos []*domain.Todo
	json.NewDecoder(rec.Body).Decode(&todos)

	if len(todos) != 3 {
		t.Errorf("expected 3 todos, got %d", len(todos))
	}
}

func TestTodoHandler_GetTodoByID(t *testing.T) {
	handler := setupTestHandler()

	// Создаем задачу
	todo := domain.Todo{Title: "Test"}
	body, _ := json.Marshal(todo)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	createRec := httptest.NewRecorder()
	handler.HandleTodos(createRec, createReq)

	var created domain.Todo
	json.NewDecoder(createRec.Body).Decode(&created)

	t.Run("получение существующей задачи", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", created.ID), nil)
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("получение несуществующей задачи", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos/9999", nil)
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("получение задачи с некорректным ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos/invalid", nil)
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestTodoHandler_UpdateTodo(t *testing.T) {
	handler := setupTestHandler()

	// Создаем задачу
	todo := domain.Todo{Title: "Original"}
	body, _ := json.Marshal(todo)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	createRec := httptest.NewRecorder()
	handler.HandleTodos(createRec, createReq)

	var created domain.Todo
	json.NewDecoder(createRec.Body).Decode(&created)

	t.Run("успешное обновление задачи", func(t *testing.T) {
		updated := domain.Todo{
			Title:       "Updated",
			Description: "New Description",
			Completed:   true,
		}

		body, _ := json.Marshal(updated)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", created.ID), bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("обновление с пустым заголовком", func(t *testing.T) {
		updated := domain.Todo{Title: ""}

		body, _ := json.Marshal(updated)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", created.ID), bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("обновление несуществующей задачи", func(t *testing.T) {
		updated := domain.Todo{Title: "Test"}

		body, _ := json.Marshal(updated)
		req := httptest.NewRequest(http.MethodPut, "/todos/9999", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestTodoHandler_DeleteTodo(t *testing.T) {
	handler := setupTestHandler()

	// Создаем задачу
	todo := domain.Todo{Title: "To Delete"}
	body, _ := json.Marshal(todo)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	createRec := httptest.NewRecorder()
	handler.HandleTodos(createRec, createReq)

	var created domain.Todo
	json.NewDecoder(createRec.Body).Decode(&created)

	t.Run("успешное удаление задачи", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/todos/%d", created.ID), nil)
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("удаление несуществующей задачи", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/todos/9999", nil)
		rec := httptest.NewRecorder()

		handler.HandleTodoByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}
