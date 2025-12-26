package domain

import (
	"context"
	"errors"
)

// Todo представляет сущность задачи
type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Validate проверяет корректность данных задачи
func (t *Todo) Validate() error {
	if t.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}

// TodoRepository определяет интерфейс для работы с хранилищем задач
type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	GetAll(ctx context.Context) ([]*Todo, error)
	GetByID(ctx context.Context, id int) (*Todo, error)
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) bool
}

// Предопределенные ошибки
var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrTodoAlreadyExists = errors.New("todo with this ID already exists")
	ErrInvalidTodoData   = errors.New("invalid todo data")
)
