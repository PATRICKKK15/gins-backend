# ГИНС-платформа — backend

Backend-часть платформы ГИНС на Go. Связывает заказчиков задач и
свободных исполнителей: заказчик размещает задачу, исполнитель на неё
откликается, заказчик принимает или отклоняет отклик.

## Стек

- Go
- PostgreSQL
- chi (роутер)
- JWT для авторизации

## Структура проекта

```
cmd/server/       — точка входа (main.go)
internal/models/  — структуры User, Task, Response
internal/repository/ — работа с базой данных
internal/auth/    — хэширование паролей и JWT
internal/middleware/ — проверка токена и роли
internal/handlers/ — HTTP-хендлеры
internal/router/  — сборка всех маршрутов API
internal/config/  — чтение .env
migrations/       — SQL-миграции схемы БД
```

## Запуск

1. Установить Go (не ниже 1.21) и PostgreSQL.

2. Создать базу данных:

   ```
   createdb gins_db
   ```

3. Применить миграции (можно просто выполнить SQL-файл руками):

   ```
   psql -d gins_db -f migrations/0001_init.sql
   ```

4. Скопировать `.env.example` в `.env` и подставить свои значения:

   ```
   cp .env.example .env
   ```

5. Установить зависимости и запустить сервер:

   ```
   go mod tidy
   go run cmd/server/main.go
   ```

6. Проверить, что сервер поднялся:

   ```
   curl http://localhost:8080/
   ```

## Основные эндпоинты

| Метод | Путь                        | Кто может вызвать       | Описание                          |
|-------|-----------------------------|-------------------------|------------------------------------|
| POST  | /api/auth/register          | любой                   | регистрация                        |
| POST  | /api/auth/login             | любой                   | вход, получение JWT                |
| GET   | /api/users/me                | любой авторизованный    | свой профиль                       |
| POST  | /api/tasks                   | заказчик                | создать задачу                     |
| GET   | /api/tasks                   | любой авторизованный    | список задач                       |
| GET   | /api/tasks/{id}               | любой авторизованный    | задача по id                        |
| POST  | /api/tasks/{id}/responses      | исполнитель             | откликнуться на задачу              |
| PATCH | /api/responses/{id}            | заказчик (владелец задачи) | принять/отклонить отклик         |

Для всех эндпоинтов, кроме `/api/auth/*`, нужен заголовок:

```
Authorization: Bearer <токен из /api/auth/login>
```

## Пример работы (curl)

```bash
# регистрация заказчика
curl -X POST localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Иван Иванов","email":"ivan@example.com","password":"password123","role":"customer"}'

# вход
curl -X POST localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ivan@example.com","password":"password123"}'

# создание задачи (подставить токен из ответа логина)
curl -X POST localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <токен>" \
  -d '{"title":"Разработать модуль X","description":"Нужно сделать модуль X для проекта Y"}'
```

## Что упрощено сознательно

Это учебный проект под задание практики, а не production-система, поэтому
намеренно не сделано:

- пагинация в списке задач (для небольшого объёма данных не критично);
- refresh-токены (JWT просто живёт сутки и всё);
- отдельный слой сервисов между хендлерами и репозиториями — для трёх
  сущностей это было бы лишней прослойкой без реальной пользы.

Если проект нужно будет расширять — это первые кандидаты на доработку.
