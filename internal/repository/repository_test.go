package repository_test

import (
	"context"
	"testing"

	"todo/internal/domain"
	"todo/internal/repository"
)

func TestInMemoryTodoRepository_Create(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	t.Run("создание задачи без ID", func(t *testing.T) {
		todo := &domain.Todo{
			Title:       "Test Todo",
			Description: "Test Description",
			Completed:   false,
		}

		err := repo.Create(ctx, todo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if todo.ID == 0 {
			t.Error("expected ID to be assigned")
		}
	})

	t.Run("создание задачи с ID", func(t *testing.T) {
		todo := &domain.Todo{
			ID:          100,
			Title:       "Test Todo with ID",
			Description: "Test Description",
			Completed:   false,
		}

		err := repo.Create(ctx, todo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if todo.ID != 100 {
			t.Errorf("expected ID to be 100, got %d", todo.ID)
		}
	})

	t.Run("создание задачи с дублирующим ID", func(t *testing.T) {
		todo1 := &domain.Todo{
			ID:    200,
			Title: "First Todo",
		}
		repo.Create(ctx, todo1)

		todo2 := &domain.Todo{
			ID:    200,
			Title: "Second Todo",
		}

		err := repo.Create(ctx, todo2)
		if err != domain.ErrTodoAlreadyExists {
			t.Errorf("expected ErrTodoAlreadyExists, got %v", err)
		}
	})
}

func TestInMemoryTodoRepository_GetAll(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	t.Run("получение пустого списка", func(t *testing.T) {
		todos, err := repo.GetAll(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(todos) != 0 {
			t.Errorf("expected empty list, got %d items", len(todos))
		}
	})

	t.Run("получение списка с задачами", func(t *testing.T) {
		repo.Create(ctx, &domain.Todo{Title: "Todo 1"})
		repo.Create(ctx, &domain.Todo{Title: "Todo 2"})

		todos, err := repo.GetAll(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(todos) != 2 {
			t.Errorf("expected 2 todos, got %d", len(todos))
		}
	})
}

func TestInMemoryTodoRepository_GetByID(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	t.Run("получение существующей задачи", func(t *testing.T) {
		original := &domain.Todo{
			Title:       "Test Todo",
			Description: "Description",
		}
		repo.Create(ctx, original)

		retrieved, err := repo.GetByID(ctx, original.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if retrieved.Title != original.Title {
			t.Errorf("expected title %s, got %s", original.Title, retrieved.Title)
		}
	})

	t.Run("получение несуществующей задачи", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 9999)
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}

func TestInMemoryTodoRepository_Update(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	t.Run("обновление существующей задачи", func(t *testing.T) {
		todo := &domain.Todo{Title: "Original"}
		repo.Create(ctx, todo)

		todo.Title = "Updated"
		todo.Completed = true

		err := repo.Update(ctx, todo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		updated, _ := repo.GetByID(ctx, todo.ID)
		if updated.Title != "Updated" || !updated.Completed {
			t.Error("todo was not updated correctly")
		}
	})

	t.Run("обновление несуществующей задачи", func(t *testing.T) {
		todo := &domain.Todo{ID: 9999, Title: "Non-existent"}
		err := repo.Update(ctx, todo)
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}

func TestInMemoryTodoRepository_Delete(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	t.Run("удаление существующей задачи", func(t *testing.T) {
		todo := &domain.Todo{Title: "To Delete"}
		repo.Create(ctx, todo)

		err := repo.Delete(ctx, todo.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = repo.GetByID(ctx, todo.ID)
		if err != domain.ErrTodoNotFound {
			t.Error("todo was not deleted")
		}
	})

	t.Run("удаление несуществующей задачи", func(t *testing.T) {
		err := repo.Delete(ctx, 9999)
		if err != domain.ErrTodoNotFound {
			t.Errorf("expected ErrTodoNotFound, got %v", err)
		}
	})
}

func TestInMemoryTodoRepository_Exists(t *testing.T) {
	repo := repository.NewInMemoryTodoRepository()
	ctx := context.Background()

	todo := &domain.Todo{Title: "Test"}
	repo.Create(ctx, todo)

	if !repo.Exists(ctx, todo.ID) {
		t.Error("expected todo to exist")
	}

	if repo.Exists(ctx, 9999) {
		t.Error("expected todo not to exist")
	}
}
