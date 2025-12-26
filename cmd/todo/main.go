package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/internal/http/handler"
	"todo/internal/http/middleware"
	"todo/internal/repository"
	"todo/internal/usecase"
)

func main() {
	// Инициализация зависимостей
	todoRepo := repository.NewInMemoryTodoRepository()
	todoUseCase := usecase.NewTodoUseCase(todoRepo)
	todoHandler := handler.NewTodoHandler(todoUseCase)

	// Настройка роутера
	mux := http.NewServeMux()

	// Регистрация эндпоинтов
	mux.HandleFunc("/todos", todoHandler.HandleTodos)
	mux.HandleFunc("/todos/", todoHandler.HandleTodoByID)

	// Применение middleware
	handlerWithMiddleware := middleware.Logger(
		middleware.Recovery(
			middleware.Timeout(30 * time.Second)(mux),
		),
	)

	// Настройка сервера
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlerWithMiddleware,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Server started successfully")

	// Ожидание сигнала для остановки
	<-done
	log.Println("Server stopping...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
