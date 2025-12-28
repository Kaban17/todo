import http from "k6/http";
import { check, group, sleep } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";

// Пользовательские метрики
const errorRate = new Rate("errors");
const todoCreationTime = new Trend("todo_creation_time");
const todoUpdateTime = new Trend("todo_update_time");
const todoDeletionTime = new Trend("todo_deletion_time");
const successfulRequests = new Counter("successful_requests");

// Конфигурация базового URL
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

// Опции для различных типов тестов
export const options = {
  stages: [
    { duration: "30s", target: 20 }, // Разогрев: постепенный рост до 20 пользователей
    { duration: "1m", target: 50 }, // Рост нагрузки до 50 пользователей
    { duration: "2m", target: 50 }, // Поддержание нагрузки 50 пользователей
    { duration: "1m", target: 100 }, // Пик: 100 пользователей
    { duration: "2m", target: 100 }, // Поддержание пика
    { duration: "1m", target: 20 }, // Снижение нагрузки
    { duration: "30s", target: 0 }, // Остывание
  ],
  thresholds: {
    http_req_duration: ["p(95)<500", "p(99)<1000"], // 95% запросов < 500ms, 99% < 1s
    http_req_failed: ["rate<0.05"], // Менее 5% ошибок
    errors: ["rate<0.05"], // Менее 5% ошибок бизнес-логики
    checks: ["rate>0.95"], // 95% проверок проходят
  },
  // Пороги для успешного прохождения теста
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

// Функция настройки теста
export function setup() {
  console.log(`Starting load test against ${BASE_URL}`);

  // Проверяем доступность API
  const res = http.get(`${BASE_URL}/todos`);
  if (res.status !== 200) {
    throw new Error(`API is not available. Status: ${res.status}`);
  }

  return { startTime: new Date().toISOString() };
}

// Главный сценарий теста
export default function (data) {
  group("Todo API Load Test", () => {
    // Сценарий 1: Создание задачи
    const createRes = createTodo();

    if (createRes.status === 201) {
      const todoId = JSON.parse(createRes.body).id;

      // Сценарий 2: Получение созданной задачи
      getTodo(todoId);

      // Сценарий 3: Получение всех задач
      getAllTodos();

      // Сценарий 4: Обновление задачи
      updateTodo(todoId);

      // Сценарий 5: Удаление задачи
      deleteTodo(todoId);
    }
  });

  // Случайная пауза между итерациями (1-3 секунды)
}

// Функция создания задачи
function createTodo() {
  const payload = JSON.stringify({
    title: `Load Test Todo ${Date.now()}`,
    description: `Created by k6 at ${new Date().toISOString()}`,
    completed: false,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
    tags: { name: "CreateTodo" },
  };

  const res = http.post(`${BASE_URL}/todos`, payload, params);

  // Проверки
  const checkResult = check(res, {
    "create todo: status is 201": (r) => r.status === 201,
    "create todo: has id": (r) => JSON.parse(r.body).id !== undefined,
    "create todo: title matches": (r) =>
      JSON.parse(r.body).title.includes("Load Test"),
    "create todo: response time < 500ms": (r) => r.timings.duration < 500,
  });

  // Обновляем метрики
  errorRate.add(!checkResult);
  todoCreationTime.add(res.timings.duration);

  if (checkResult) {
    successfulRequests.add(1);
  }

  return res;
}

// Функция получения задачи по ID
function getTodo(todoId) {
  const params = {
    tags: { name: "GetTodo" },
  };

  const res = http.get(`${BASE_URL}/todos/${todoId}`, params);

  const checkResult = check(res, {
    "get todo: status is 200": (r) => r.status === 200,
    "get todo: has correct id": (r) => JSON.parse(r.body).id === todoId,
    "get todo: response time < 300ms": (r) => r.timings.duration < 300,
  });

  errorRate.add(!checkResult);

  if (checkResult) {
    successfulRequests.add(1);
  }

  return res;
}

// Функция получения всех задач
function getAllTodos() {
  const params = {
    tags: { name: "GetAllTodos" },
  };

  const res = http.get(`${BASE_URL}/todos`, params);

  const checkResult = check(res, {
    "get all todos: status is 200": (r) => r.status === 200,
    "get all todos: returns array": (r) => Array.isArray(JSON.parse(r.body)),
    "get all todos: response time < 400ms": (r) => r.timings.duration < 400,
  });

  errorRate.add(!checkResult);

  if (checkResult) {
    successfulRequests.add(1);
  }

  return res;
}

// Функция обновления задачи
function updateTodo(todoId) {
  const payload = JSON.stringify({
    title: `Updated Todo ${Date.now()}`,
    description: "Updated by k6 load test",
    completed: true,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
    tags: { name: "UpdateTodo" },
  };

  const res = http.put(`${BASE_URL}/todos/${todoId}`, payload, params);

  const checkResult = check(res, {
    "update todo: status is 200": (r) => r.status === 200,
    "update todo: completed is true": (r) =>
      JSON.parse(r.body).completed === true,
    "update todo: response time < 500ms": (r) => r.timings.duration < 500,
  });

  errorRate.add(!checkResult);
  todoUpdateTime.add(res.timings.duration);

  if (checkResult) {
    successfulRequests.add(1);
  }

  return res;
}

// Функция удаления задачи
function deleteTodo(todoId) {
  const params = {
    tags: { name: "DeleteTodo" },
  };

  const res = http.del(`${BASE_URL}/todos/${todoId}`, null, params);

  const checkResult = check(res, {
    "delete todo: status is 204": (r) => r.status === 204,
    "delete todo: response time < 300ms": (r) => r.timings.duration < 300,
  });

  errorRate.add(!checkResult);
  todoDeletionTime.add(res.timings.duration);

  if (checkResult) {
    successfulRequests.add(1);
  }

  return res;
}

// Функция завершения теста
export function teardown(data) {
  console.log(`\nLoad test completed. Started at: ${data.startTime}`);
  console.log(`Finished at: ${new Date().toISOString()}`);
}
