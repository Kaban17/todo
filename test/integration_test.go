package test_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"
	"todo/internal/domain"
	"todo/internal/http/handler"
	"todo/internal/http/middleware"
	"todo/internal/repository"
	"todo/internal/usecase"
)

// setupTestServer создает тестовый HTTP сервер
func setupTestServer() http.Handler {
	repo := repository.NewInMemoryTodoRepository()
	uc := usecase.NewTodoUseCase(repo)
	h := handler.NewTodoHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("/todos", h.HandleTodos)
	mux.HandleFunc("/todos/", h.HandleTodoByID)

	// Применяем middleware
	return middleware.Logger(
		middleware.Recovery(
			middleware.Timeout(5 * time.Second)(mux),
		),
	)
}

func TestIntegration_FullTodoLifecycle(t *testing.T) {
	server := setupTestServer()

	// 1. Создаем задачу
	createTodo := domain.Todo{
		Title:       "Integration Test Todo",
		Description: "Test description",
		Completed:   false,
	}

	body, _ := json.Marshal(createTodo)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", rec.Code)
	}

	var created domain.Todo
	json.NewDecoder(rec.Body).Decode(&created)

	if created.ID == 0 {
		t.Fatal("Expected ID to be assigned")
	}

	// 2. Получаем созданную задачу
	req = httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var retrieved domain.Todo
	json.NewDecoder(rec.Body).Decode(&retrieved)

	if retrieved.Title != createTodo.Title {
		t.Errorf("Expected title %s, got %s", createTodo.Title, retrieved.Title)
	}

	// 3. Обновляем задачу
	updateTodo := domain.Todo{
		Title:       "Updated Todo",
		Description: "Updated description",
		Completed:   true,
	}

	body, _ = json.Marshal(updateTodo)
	req = httptest.NewRequest(http.MethodPut, "/todos/1", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var updated domain.Todo
	json.NewDecoder(rec.Body).Decode(&updated)

	if updated.Title != updateTodo.Title {
		t.Errorf("Expected title %s, got %s", updateTodo.Title, updated.Title)
	}

	if !updated.Completed {
		t.Error("Expected todo to be completed")
	}

	// 4. Получаем все задачи
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var todos []*domain.Todo
	json.NewDecoder(rec.Body).Decode(&todos)

	if len(todos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(todos))
	}

	// 5. Удаляем задачу
	req = httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", rec.Code)
	}

	// 6. Проверяем, что задача удалена
	req = httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", rec.Code)
	}
}

func TestIntegration_MultipleTodos(t *testing.T) {
	server := setupTestServer()

	// Создаем несколько задач
	todos := []domain.Todo{
		{Title: "Todo 1", Description: "First", Completed: false},
		{Title: "Todo 2", Description: "Second", Completed: true},
		{Title: "Todo 3", Description: "Third", Completed: false},
	}

	for _, todo := range todos {
		body, _ := json.Marshal(todo)
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Failed to create todo: %s", todo.Title)
		}
	}

	// Получаем все задачи
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var allTodos []*domain.Todo
	json.NewDecoder(rec.Body).Decode(&allTodos)

	if len(allTodos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(allTodos))
	}
}

func TestIntegration_ValidationErrors(t *testing.T) {
	server := setupTestServer()

	testCases := []struct {
		name           string
		todo           domain.Todo
		expectedStatus int
	}{
		{
			name:           "empty title",
			todo:           domain.Todo{Title: "", Description: "Test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid todo",
			todo:           domain.Todo{Title: "Valid", Description: "Test"},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.todo)
			req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}

func TestIntegration_NotFoundErrors(t *testing.T) {
	server := setupTestServer()

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "get non-existent todo",
			method: http.MethodGet,
			path:   "/todos/999",
		},
		{
			name:   "update non-existent todo",
			method: http.MethodPut,
			path:   "/todos/999",
		},
		{
			name:   "delete non-existent todo",
			method: http.MethodDelete,
			path:   "/todos/999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			if tc.method == http.MethodPut {
				body, _ := json.Marshal(domain.Todo{Title: "Test"})
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("Expected status 404, got %d", rec.Code)
			}
		})
	}
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	server := setupTestServer()

	// Создаем задачи конкурентно
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			todo := domain.Todo{
				Title:       "Concurrent Todo",
				Description: "Test concurrency",
				Completed:   false,
			}

			body, _ := json.Marshal(todo)
			req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("Failed to create todo concurrently")
			}

			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Проверяем, что все задачи созданы
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	var todos []*domain.Todo
	json.NewDecoder(rec.Body).Decode(&todos)

	if len(todos) != numGoroutines {
		t.Errorf("Expected %d todos, got %d", numGoroutines, len(todos))
	}
}
