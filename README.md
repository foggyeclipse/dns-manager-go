# DNS Manager

Клиент-серверное приложение для управления DNS-серверами на удалённой машине через файл `/etc/resolv.conf`.

## Описание

Приложение позволяет добавлять, удалять и просматривать список DNS-серверов (nameservers). 

**Сервер** предоставляет REST API, а **CLI-клиент** обеспечивает управление через командную строку.

## Основные возможности

- Добавление и удаление nameserver
- Получение актуального списка DNS-серверов
- Валидация IP-адресов
- Атомарное обновление конфигурации с использованием file locking

## Структура проекта

```bash
dns-manager-go/
├── go.mod
├── .golangci.yml
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   ├── server/
│   │   └── main.go          # HTTP-сервер на Gin
│   └── client/
│       └── main.go          # CLI-клиент на Cobra
├── internal/
│   ├── dnsmanager/
│   │   ├── manager.go       # Логика работы с /etc/resolv.conf
│   │   └── manager_test.go  # Unit-тесты менеджера
│   └── handler/
│       └── dns.go           # HTTP-обработчики
└── pkg/
    └── validator/
        └── validator.go     # Валидация IP-адресов
```

## Технологии

- **Язык**: Go 1.23
- **Фреймворки**: Gin, Cobra
- **Тестирование**: testify
- **CI/CD**: GitHub Actions + golangci-lint

## Требования

- Go 1.23+
- Linux (поддерживается работа с `/etc/resolv.conf` и `flock`)
- Права суперпользователя для запуска сервера

## Сборка и запуск

### Зависимости

```bash
go mod tidy
go mod verify
```

### Сборка

```bash
# Вариант 1: Полная сборка проекта (проверка компиляции)
go build ./...

# Вариант 2: Сборка исполняемых файлов
go build -o bin/dns-server ./cmd/server
go build -o bin/dns-client ./cmd/client
```

### Запуск

**Сервер:**

```bash
# Вариант 1: Запуск через go run (удобно для разработки)
sudo go run cmd/server/main.go
sudo go run cmd/server/main.go --port 8080

# Вариант 2: Запуск собранного бинарника
sudo ./bin/dns-server
sudo ./bin/dns-server --port 8080
```

**Клиент:**

```bash
# Справка по командам
go run cmd/client/main.go --help

# Получить список nameservers
go run cmd/client/main.go --server http://localhost:8080 list

# Добавить nameserver
go run cmd/client/main.go --server http://localhost:8080 add 8.8.8.8

# Удалить nameserver
go run cmd/client/main.go --server http://localhost:8080 remove 1.1.1.1
```

## Тестирование и проверка качества

```bash
# Запуск всех тестов
go test ./... -v

# Тесты менеджера с покрытием
go test ./internal/dnsmanager -v -cover

# Проверка качества кода
go fmt ./...
go vet ./...
golangci-lint run
```

## API

| Метод   | Эндпоинт   | Описание                     |
|---------|------------|------------------------------|
| `GET`   | `/dns`     | Получить список nameservers  |
| `POST`  | `/dns`     | Добавить nameserver          |
| `DELETE`| `/dns`     | Удалить nameserver           |
| `GET`   | `/health`  | Проверка состояния сервиса   |
