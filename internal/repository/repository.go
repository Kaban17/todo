package repository

import (
	"context"
	"sync"

	"todo/internal/domain"
)

// InMemoryTodoRepository реализует хранилище задач в памяти
type InMemoryTodoRepository struct {
	mu     sync.RWMutex
	todos  map[int]*domain.Todo
	nextID int
}

// NewInMemoryTodoRepository создает новый экземпляр репозитория
func NewInMemoryTodoRepository() *InMemoryTodoRepository {
	return &InMemoryTodoRepository{
		todos:  make(map[int]*domain.Todo),
		nextID: 1,
	}
}

// Create создает новую задачу
func (r *InMemoryTodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Если ID не указан, генерируем новый
	if todo.ID == 0 {
		todo.ID = r.nextID
		r.nextID++
	} else {
		// Проверяем, не существует ли уже задача с таким ID
		if _, exists := r.todos[todo.ID]; exists {
			return domain.ErrTodoAlreadyExists
		}
		// Обновляем nextID если нужно
		if todo.ID >= r.nextID {
			r.nextID = todo.ID + 1
		}
	}

	r.todos[todo.ID] = todo
	return nil
}

// GetAll возвращает все задачи
func (r *InMemoryTodoRepository) GetAll(ctx context.Context) ([]*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todos := make([]*domain.Todo, 0, len(r.todos))
	for _, todo := range r.todos {
		todos = append(todos, todo)
	}

	return todos, nil
}

// GetByID возвращает задачу по идентификатору
func (r *InMemoryTodoRepository) GetByID(ctx context.Context, id int) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, exists := r.todos[id]
	if !exists {
		return nil, domain.ErrTodoNotFound
	}

	return todo, nil
}

// Update обновляет существующую задачу
func (r *InMemoryTodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[todo.ID]; !exists {
		return domain.ErrTodoNotFound
	}

	r.todos[todo.ID] = todo
	return nil
}

// Delete удаляет задачу по идентификатору
func (r *InMemoryTodoRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[id]; !exists {
		return domain.ErrTodoNotFound
	}

	delete(r.todos, id)
	return nil
}

// Exists проверяет существование задачи
func (r *InMemoryTodoRepository) Exists(ctx context.Context, id int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.todos[id]
	return exists
}
