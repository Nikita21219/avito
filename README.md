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
- redis - выступает в роли хранения ключей идемпотентности, хранения кеша активных сегментов пользователя, а также хранения task_id
- swagger - отображает документацию OpenAPI на порту 8081

## Реализовано

- Проверка ключа идемпотентности для методов которые меняют состояние сервера
- Кеширование активных сегменов пользователей
- Rate limiter
- Документация каждой функции
- Покрытие unit тестами
- Логирование
- Доп. задание 1: формаирование csv отчета по пользователю за определенный период
- Доп. задание 2: возможность задавать TTL (время автоматического удаления пользователя из сегмента)

## Генерация отчетов 
Генерация отчетов может занимать много времени, поэтому я принял решение генерировать отчет в отдельной горутине используя контекст с тайм-аутом.
Для этого потребовалось создать два ендпоинта - один для запуска генерации отчета и второй для проверки результата и получения ссылки на скачивание файла

## Время автоматического удаления пользователя из сегмента
Минимальная нагрузка сервера обычно ночью, поэтому каждый день в 03:00 по МСК или 00:00 UTC будет запускаться задача на удаление пользователей из сегментов.
В таблице user_segments столбец alive_until хранит информацию о том, до какого времени будет жить объект.
Если текущее время стало больше либо равно чем alive_until, то пользователь будет удален из сегмента

## Примеры curl запросов

Создание сегмента POST /segment
``` bash
curl -X POST "http://localhost:8080/segment" \
     -H "Content-Type: application/json" \
     -H "Idempotency-Key: unique_key_1" \
     -d '{
          "slug": "AVITO_VOICE_MESSAGES"
     }' \
     -w "%{http_code}\n"
```

Удаление сегмента DELETE /segment
``` bash
curl -X DELETE "http://localhost:8080/segment" \
     -H "Content-Type: application/json" \
     -H "Idempotency-Key: unique_key_2" \
     -d '{
          "slug": "AVITO_VOICE_MESSAGES"
     }' \
     -w "%{http_code}\n"
```

Добавление и удаление сегментов пользователя POST /segment/user
``` bash
curl -X POST "http://localhost:8080/segment/user" \
     -H "Content-Type: application/json" \
     -H "Idempotency-Key: unique_key_3" \
     -d '{
          "user_id": 1,
          "add": [
            "AVITO_VOICE_MESSAGES"
          ],
          "del": [
            "AVITO_DISCOUNT_50"
          ],
          "ttl_days": 5
     }' \
     -w "%{http_code}\n"
```

Получение сегментов пользователя GET /segment/user
``` bash
curl -X GET "http://localhost:8080/segment/user?id=1" \
     -H "Content-Type: application/json" \
     -w "%{http_code}\n"
```

Генерация отчета GET /report
``` bash
curl -X GET -G -d 'date=2023-08-29%2010:32' "http://localhost:8080/report"
```

Проверка статуса задачи по генерации отчета GET /report_check
``` bash
curl -X GET -G -d 'task_id=<id задачи с прошлого запроса>' "http://localhost:8080/report_check"
```
