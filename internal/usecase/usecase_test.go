package usecase

import (
	"context"
	"testing"

	"todo/internal/domain"
	"todo/internal/repository"
)

func TestTodoUseCase_CreateTodo(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	uc := NewTodoUseCase(repo)
	ctx := context.Background()

	t.Run("успешное создание задачи", func(t *testing.T) {
		todo := &domain.Todo{
			Title:       "New Todo",
			Description: "Description",
			Completed:   false,
		}

		created, err := uc.CreateTodo(ctx, todo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if created.ID == 0 {
			t.Error("expected ID to be assigned")
		}

		if created.Title != todo.Title {
			t.Errorf("expected title %s, got %s", todo.Title, created.Title)
		}
	})

	t.Run("создание задачи с пустым заголовком", func(t *testing.T) {
		todo := &domain.Todo{
			Title:       "",
			Description: "Description",
		}

		_, err := uc.CreateTodo(ctx, todo)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("создание задачи с дублирующим ID", func(t *testing.T) {
		todo1 := &domain.Todo{
			ID:    500,
			Title: "First",
		}
		uc.CreateTodo(ctx, todo1)

		todo2 := &domain.Todo{
			ID:    500,
			Title: "Second",
		}
		_, err := uc.CreateTodo(ctx, todo2)
		if err != domain.ErrTodoAlreadyExists {
			t.Errorf("expected ErrTodoAlreadyExists, got %v", err)
		}
	})
}

func TestTodoUseCase_GetAllTodos(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	uc := NewTodoUseCase(repo)
	ctx := context.Background()

	// Создаем несколько задач
	uc.CreateTodo(ctx, &domain.Todo{Title: "Todo 1"})
	uc.CreateTodo(ctx, &domain.Todo{Title: "Todo 2"})
	uc.CreateTodo(ctx, &domain.Todo{Title: "Todo 3"})

	todos, err := uc.GetAllTodos(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(todos) != 3 {
		t.Errorf("expected 3 todos, got %d", len(todos))
	}
}

func TestTodoUseCase_GetTodoByID(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	uc := NewTodoUseCase(repo)
	ctx := context.Background()

	t.Run("получение существующей задачи", func(t *testing.T) {
		created, _ := uc.CreateTodo(ctx, &domain.Todo{Title: "Test"})

		retrieved, err := uc.GetTodoByID(ctx, created.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, retrieved.ID)
		}
	})

	t.Run("получение несуществующей задачи", func(t *testing.T) {
		_, err := uc.GetTodoByID(ctx, 9999)
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}

func TestTodoUseCase_UpdateTodo(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	uc := NewTodoUseCase(repo)
	ctx := context.Background()

	t.Run("успешное обновление задачи", func(t *testing.T) {
		created, _ := uc.CreateTodo(ctx, &domain.Todo{Title: "Original"})

		updated, err := uc.UpdateTodo(ctx, created.ID, &domain.Todo{
			Title:       "Updated",
			Description: "New Description",
			Completed:   true,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if updated.Title != "Updated" {
			t.Error("title was not updated")
		}

		if !updated.Completed {
			t.Error("completed status was not updated")
		}
	})

	t.Run("обновление с пустым заголовком", func(t *testing.T) {
		created, _ := uc.CreateTodo(ctx, &domain.Todo{Title: "Original"})

		_, err := uc.UpdateTodo(ctx, created.ID, &domain.Todo{
			Title: "",
		})

		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("обновление несуществующей задачи", func(t *testing.T) {
		_, err := uc.UpdateTodo(ctx, 9999, &domain.Todo{Title: "Test"})
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}

func TestTodoUseCase_DeleteTodo(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	uc := NewTodoUseCase(repo)
	ctx := context.Background()

	t.Run("успешное удаление задачи", func(t *testing.T) {
		created, _ := uc.CreateTodo(ctx, &domain.Todo{Title: "To Delete"})

		err := uc.DeleteTodo(ctx, created.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = uc.GetTodoByID(ctx, created.ID)
		if err != domain.ErrTodoNotFound {
			t.Error("todo was not deleted")
		}
	})

	t.Run("удаление несуществующей задачи", func(t *testing.T) {
		err := uc.DeleteTodo(ctx, 9999)
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}
