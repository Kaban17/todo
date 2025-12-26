package usecase

import (
	"context"

	"todo/internal/domain"
)

// TodoUseCase содержит бизнес-логику для работы с задачами
type TodoUseCase struct {
	repo domain.TodoRepository
}

// NewTodoUseCase создает новый экземпляр use case
func NewTodoUseCase(repo domain.TodoRepository) *TodoUseCase {
	return &TodoUseCase{
		repo: repo,
	}
}

// CreateTodo создает новую задачу
func (uc *TodoUseCase) CreateTodo(ctx context.Context, todo *domain.Todo) (*domain.Todo, error) {
	// Валидация
	if err := todo.Validate(); err != nil {
		return nil, err
	}

	// Создание
	if err := uc.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// GetAllTodos возвращает все задачи
func (uc *TodoUseCase) GetAllTodos(ctx context.Context) ([]*domain.Todo, error) {
	return uc.repo.GetAll(ctx)
}

// GetTodoByID возвращает задачу по идентификатору
func (uc *TodoUseCase) GetTodoByID(ctx context.Context, id int) (*domain.Todo, error) {
	return uc.repo.GetByID(ctx, id)
}

// UpdateTodo обновляет существующую задачу
func (uc *TodoUseCase) UpdateTodo(ctx context.Context, id int, todo *domain.Todo) (*domain.Todo, error) {
	// Валидация
	if err := todo.Validate(); err != nil {
		return nil, err
	}

	// Проверка существования
	if !uc.repo.Exists(ctx, id) {
		return nil, domain.ErrTodoNotFound
	}

	// Установка ID
	todo.ID = id

	// Обновление
	if err := uc.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// DeleteTodo удаляет задачу
func (uc *TodoUseCase) DeleteTodo(ctx context.Context, id int) error {
	return uc.repo.Delete(ctx, id)
}
