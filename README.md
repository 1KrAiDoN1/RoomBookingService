# Room Booking Service

Сервис для бронирования переговорных комнат с гибким управлением расписанием и автоматической генерацией слотов.

[![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://www.postgresql.org/)


## 🎯 Описание проекта

Room Booking Service — это микросервис для управления бронированием переговорных комнат. Сервис позволяет администраторам создавать комнаты и настраивать их расписание, а пользователям — просматривать доступные слоты и создавать бронирования.

### Основные возможности

**Администратор:**
- ✅ Создание переговорных комнат
- ✅ Создание расписания работы комнаты (один раз, без возможности изменения)
- ✅ Просмотр всех бронирований с пагинацией

**Пользователь:**
- ✅ Просмотр списка комнат
- ✅ Просмотр доступных слотов по дате
- ✅ Создание бронирования на свободный слот
- ✅ Отмена своих бронирований (идемпотентно)
- ✅ Просмотр своих будущих бронирований

## 📊 Бизнес-требования

### Ключевые ограничения
- Слоты генерируются автоматически на основе расписания (длительность 30 минут)
- Один слот может быть занят только одной активной бронью
- Нельзя создать бронь на прошедший слот
- Отмена брони идемпотентна (повторная отмена не вызывает ошибку)
- Администратор не может создавать бронирования
- Для комнаты без расписания слоты не создаются
- Все даты и время хранятся и передаются в UTC

### Производительность
- **RPS:** 100 запросов/сек
- **SLI успешности:** 99.9%
- **SLI времени ответа:** < 200ms для эндпоинта получения слотов
- **Объем данных:** до 50 комнат, до 1k слотов/день, до 10k пользователей, до 100k броней

## 🛠 Технологический стек

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Язык | Go | 1.25.4 |
| Web-фреймворк | Gin | latest |
| База данных | PostgreSQL | 16 |
| Драйвер БД | pgx/v5 | latest |
| Миграции | golang-migrate | v4 |
| Логирование | Zap | latest |
| Конфигурация | cleanenv | latest |
| Транзакции | go-transaction-manager | latest |
| JWT | golang-jwt/jwt | v5 |
| Валидация | go-playground/validator | v10 |
| Тестирование | testify | v1.11.1 |
| Swagger | swaggo/swag | latest |
| Линтер | golangci-lint | v1.60+ |
| Контейнеризация | Docker, Docker Compose | latest |

## 🏗 Архитектура

### Clean Architecture

Проект построен на принципах Clean Architecture с четким разделением ответственности:

### Слои приложения

1. **Handler Layer** (`http-server/handler/`)
   - Обработка HTTP запросов
   - Валидация входных данных
   - Формирование ответов

2. **Service Layer** (`service/`)
   - Реализация бизнес-логики
   - Управление транзакциями
   - Обработка бизнес-ошибок

3. **Repository Layer** (`repository/postgres/`)
   - Выполнение SQL запросов
   - Работа с транзакциями через getter
   - Преобразование данных

4. **Domain Layer** (`domain/entity/`)
   - Доменные сущности
   - Бизнес-правила
   - Константы

### Генерация слотов

**Выбранный подход:** Слоты генерируются **на лету при запросе** и сохраняются в БД через `UPSERT`. Это решение обеспечивает:

- **Гибкость:** Нет необходимости генерировать слоты на длительный период вперед
- **Производительность:** Слоты генерируются только для запрашиваемых дат
- **Согласованность:** Сохранение в БД позволяет ссылаться на слоты по ID при бронировании

### Управление транзакциями

Для обеспечения целостности данных используется `go-transaction-manager`:
- Автоматическая привязка транзакции к контексту
- Поддержка вложенных транзакций
- Единый интерфейс для работы с транзакциями в репозиториях

## 🚀 Установка и запуск

### Предварительные требования

- Docker 24.0+
- Docker Compose 2.20+
- Make (опционально)
- Go 1.25.4+ (для локальной разработки)

## 🚀 Быстрый старт

### Предварительные требования

- Docker и Docker Compose

### Запуск сервиса

1. **Клонируйте репозиторий**
```bash
git clone https://github.com/avito-internships/test-backend-1-1KrAiDoN1.git
cd test-backend-1-1KrAiDoN1
```

2. **Запустите сервис**
```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`

## Примеры запросов

**Получить токен пользователя**
```bash
curl -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"user"}'
```

**Получить токен администратора**
```bash
curl -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
```
**Работа с комнатами**

```bash
# Создание комнаты (admin)
curl -X POST http://localhost:8080/rooms/create \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Конференц-зал Москва",
    "description": "Большой зал для конференций",
    "capacity": 50
  }'
```
```bash
# Список комнат
curl -X GET http://localhost:8080/rooms/list \
  -H "Authorization: Bearer <USER_TOKEN>"
```
**Работа с расписанием**
```bash
# Создание расписания (admin)
curl -X POST http://localhost:8080/rooms/<ROOM_ID>/schedule/create \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "daysOfWeek": [1,2,3,4,5],
    "startTime": "09:00",
    "endTime": "18:00"
  }'
  ```

**Получение слотов**
```bash
# Получить слоты на завтра
TOMORROW=$(date -u -v+1d +"%Y-%m-%d")
curl -X GET "http://localhost:8080/rooms/<ROOM_ID>/slots/list?date=$TOMORROW" \
  -H "Authorization: Bearer <USER_TOKEN>"
  ```
**Работа с бронированиями**
```bash
# Создание брони
curl -X POST http://localhost:8080/bookings/create \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "slotId": "<SLOT_ID>",
    "createConferenceLink": false
  }'
```
```bash
# Мои брони
curl -X GET http://localhost:8080/bookings/my \
  -H "Authorization: Bearer <USER_TOKEN>"

# Отмена брони
curl -X POST http://localhost:8080/bookings/<BOOKING_ID>/cancel \
  -H "Authorization: Bearer <USER_TOKEN>"

# Все брони (admin)
curl -X GET "http://localhost:8080/bookings/list?page=1&pageSize=20" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

## 🛠 Конфигурация инструментов

### Линтер (golangci-lint)

Файл конфигурации линтера `.golangci.yml`:

```yaml
version: "2"
run:
  timeout: 3m
  tests: false 

linters:
  disable-all: true
  enable:
    - errcheck      # Проверка обработки ошибок
    - govet         # Анализ кода на подозрительные конструкции
    - ineffassign   # Поиск неэффективных присваиваний
    - staticcheck   # Статический анализ (замена golint)

issues:
  exclude-use-default: true
```
## MakeFile
```bash
# Запуск линтера
make lint

# Сборка Docker образов
make docker-build

# Запуск всех сервисов
make docker-up

# Остановка сервисов
make docker-down

# Просмотр логов
make docker-logs

# Применение миграций
make migrate-up

# Откат миграций
make migrate-down
```

## 📚 API Документация

**Swagger UI:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/xR-tWBKa)