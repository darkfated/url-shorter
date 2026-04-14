# url-shorter

Сервис по сокращению ссылок. Позволяет делиться контентом быстрее и лаконичнее: пользователь предоставляет ссылку, а взамен получает сокращённый вариант.

## 📃 Использование

### Запуск через Docker

Компиляция:
```bash
docker compose build
```

Запуск:
```bash
docker compose up -d
```

Остановка:
```bash
docker compose down
```

### Локальный запуск

Запуск:
```bash
go run ./cmd/url-shorter
```

Тестирование:
```bash
go test ./...
```

### Конфигурация

Имеется два варианта хранилища:
- `memory` - сохраняет данные на время исполнения приложения
- `postgres` - взаимодействует с базой данных

**Перед применением сервиса, создайте и заполните конфиг-файл:**
```bash
cp .env.example .env
```

## 🧩 Эндпоинты

### Отправка ссылки

Запрос:
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"http://yandex.ru\"}"
```

Ответ:
```json
{
  "original_url": "http://yandex.ru",
  "short_url": "http://localhost:8080/YjCYGkZ1O0",
  "short_code": "YjCYGkZ1O0"
}
```

### Переход по сокращённой ссылке

Запрос:
```bash
curl -i http://localhost:8080/YjCYGkZ1O0
```

Ответ (пересылает на сайт):

```text
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: http://yandex.ru
Date: Wed, 15 Apr 2026 0:44:21 GMT
Content-Length: 39

<a href="http://yandex.ru">Found</a>.
```

### Проверка работоспособности

Запрос:
```bash
curl http://localhost:8080/healthz
```

Ответ:
```text
ok
```

## 🔒 Лицензия

Этот проект лицензирован под MIT License. [MIT License](https://github.com/darkfated/url-shorter/blob/master/LICENSE)
