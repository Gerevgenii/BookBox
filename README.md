# BookBox

BookBox — микросервис для управления книгами и авторами с поддержкой миграций Goose, gRPC и паттерна Outbox для гарантированной доставки сообщений.

### Ключевые фичи

1) PostgreSQL + UUID + триггеры для created_at/updated_at.
2) Goose‑миграции с go:embed.
3) gRPC API + HTTP Gateway.
4) Outbox pattern для асинхронных веб‑хуков.
5) Конфиг через env‑переменные, Docker Compose.

### Запуск
```bash
docker-compose up -d
make generate
make build
go run ./cmd/migrate
go run ./cmd/server
```

### На этапе создания
- **Логирование → Loki + Promtail**  
  - Конфиг для Promtail: сбор JSON‑логов из файла `/var/log/library.log`, вычленение меток (`level`, `trace_id`, `book_id`, `author_id`, `component`).  
  - Настроить Grafana с datasource для Loki и простые дэшборды по логам (по уровню, trace_id).

- **Трассировка → Jaeger + OpenTelemetry**  
  - Внедрить gRPC‑интерсепторы `otelgrpc` для автоматического создания span’ов в gRPC‑сервисе.  
  - Добавить атрибуты (`span.SetAttributes`) с ключевыми параметрами (например, `book.id`).  
  - Проверить, что в Jaeger отображаются запросы с trace_id.

- **Метрики → Prometheus + Grafana**  
  - Вставить `promhttp.Handler()` на `/metrics` (порт из конфига).  
  - Конфиг Prometheus: скрейпить `library:9000` и `postgres:9187`.  
  - В Grafana добавить дашборды:
    - Outbox: tasks by kind, processing rate, error rate.
    - API: RPS и latency по эндпоинтам, `go_goroutines`, `heap_inuse_bytes`.
    - Postgres: row count и insert rate по таблицам.

- **Генерация нагрузки**  
  - Написать сценарий на k6 для создания/чтения книг и авторов.  
  - Прогонять сценарий перед демо, чтобы заполнить метрики и логи.
