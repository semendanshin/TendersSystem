# API системы тендеров

## Запуск проекта

* ### Docker

1. Создайте файл `.env` в папке `database` и укажите в нем переменные окружения:

    ```bash
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=postgres
    POSTGRES_DB=postgres
    ```

2. Создайте файл `.env` в корне проекта и укажите в нем переменные окружения:

    ```bash
    SERVER_ADDRESS="0.0.0.0:8080"
   POSTGRES_USERNAME="postgres"
   POSTGRES_PASSWORD="postgres"
   POSTGRES_DATABASE="postgres"
   POSTGRES_HOST="localhost"
   POSTGRES_PORT="5432"
    ```
   
3. Запустите проект:

    ```bash
    make docker_with_db
    ```
   
4. Примените миграции:

    ```bash
    make migrate_up DATABASE_URL="host={host} port={port} user={user} password={password} dbname={dbname} sslmode=disable"
    ```
   
5. Готово!

