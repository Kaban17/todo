FROM golang:1.25-alpine

WORKDIR /app


# Копируем файлы модулей
COPY go.mod  ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o myapp ./cmd/todo

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./myapp"]
