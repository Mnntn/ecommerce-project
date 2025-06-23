# E-commerce Microservices Project

## Описание

Микросервисная e-commerce система с асинхронной обработкой заказов, реализованная на Go (backend) и React (frontend). Проект разворачивается через Docker Compose, использует Kafka для событий и Postgres для хранения данных. Включает автоматизированную документацию (Swagger) и коллекцию Postman.

---

## Архитектура

- **Order Service** — создание и хранение заказов, публикация событий в Kafka, transactional outbox.
- **Payment Service** — управление счетами пользователей, пополнение, списание, transactional inbox/outbox, асинхронная обработка платежей.
- **API Gateway** — проксирование HTTP-запросов к backend-сервисам, CORS, единая точка входа.
- **Frontend** — React-приложение (корзина, заказы, аккаунт), интеграция через API Gateway.
- **Postgres** — хранение данных.
- **Kafka/Zookeeper** — асинхронные события между сервисами.

---

## Быстрый старт

1. **Запуск всех сервисов:**
   ```bash
   docker compose up --build
   ```
2. **Проверка доступности:**
   - Frontend: [http://localhost:3000](http://localhost:3000)
   - API Gateway: [http://localhost:8080](http://localhost:8080)
   - Swagger UI: [http://localhost:8083](http://localhost:8083)

3. **Документация и тестирование:**
   - Swagger/OpenAPI: `swagger/api-gateway-swagger.yaml` ([swagger/README.md](./swagger/README.md))
   - Postman: `swagger/ecommerce-postman-collection.json`

---

## Критерии и ответы на вопросы задания

- **Асинхронная обработка заказов:**
  - Заказ создаётся со статусом NEW, событие отправляется в Kafka.
  - Payment Service асинхронно обрабатывает событие, списывает деньги, отправляет статус заказа обратно через Kafka.
  - Order Service обновляет статус заказа на FINISHED или CANCELLED.
- **Transactional Outbox/Inbox:**
  - Order Service: заказ и событие пишутся в одной транзакции, отдельный процессор отправляет события в Kafka.
  - Payment Service: входящее событие сохраняется в inbox, обработка и публикация статуса заказа происходят в одной транзакции, отдельный процессор отправляет события из outbox в Kafka.
  - Гарантируется exactly-once семантика при списании денег.
- **CORS:**
  - В API Gateway реализован middleware, который всегда добавляет CORS-заголовки для всех ответов.
- **Документация:**
  - Swagger спецификация и коллекция Postman включены, Swagger UI доступен через Docker.
- **Автоматизация:**
  - Весь проект разворачивается одной командой через Docker Compose.
- **Чистота кода:**
  - Dead code и устаревшие файлы удалены, структура микросервисов прозрачна.

---

## Как устроен код

- **Order Service**
  - `internal/domain` — бизнес-модели (Order, Product, Events)
  - `internal/service` — бизнес-логика, обработка заказов, outbox
  - `internal/kafka` — работа с Kafka
  - `internal/repository` — доступ к БД (Postgres)
  - `internal/transport/http` — HTTP-обработчики
- **Payment Service**
  - `internal/domain` — Account, User, Events
  - `internal/service` — бизнес-логика, обработка платежей, inbox/outbox
  - `internal/kafka` — работа с Kafka
  - `internal/repository` — доступ к БД (Postgres)
  - `internal/transport/http` — HTTP-обработчики
- **API Gateway**
  - `internal/router` — маршрутизация и проксирование
  - `internal/middleware` — CORS, (Auth)
- **Frontend**
  - `src/pages` — основные страницы (Home, Cart, Orders, Account)
  - `src/context` — контекст корзины
  - `src/components` — компоненты UI

---

## Типовые вопросы

- **Как добавить новый endpoint?**
  - Добавьте маршрут в API Gateway (`internal/router/router.go`), реализуйте обработку в нужном сервисе, обновите Swagger.
- **Как добавить бизнес-логику?**
  - В соответствующий сервис: `internal/service` и/или `internal/handler`.
- **Как тестировать?**
  - Через Swagger UI (`http://localhost:8083`) или Postman (коллекция в swagger/).
- **Как отлаживать?**
  - Логи сервисов: `docker compose logs <service>`
  - Проверка Kafka: см. [swagger/README.md](./swagger/README.md)

---
