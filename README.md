# Сервис динамического сегментирования пользователей Avito

- Свагер документация: http://localhost:8081
- Приложение: http://localhost:8080

Usage:
``` bash
docker-compose up -d --build
```
Test coverage:
``` bash
cd app; go test -coverpkg=./cmd/web/handlers/... -coverprofile=coverage.out ./tests -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && rm coverage.out
```

## Сервисы докер
 
- app - go приложение
- db - база данных PostgreSQL
- redis - выступает в роли хранения ключей идемпотентности методов а также хранения кеша
- swagger - отображает документацию OpenAPI на порту 8081

## Реализовано

- Проверка ключа идемпотентности для методов которые меняют состояние сервера
- Кеширование активных сегменов пользователей
- Rate limiter
- Документация каждой функции
- Покрытие unit тестами (coverage: 76.4%)
- Логирование
