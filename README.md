# API системы тендеров

> Честно отмечу, что выполнение задания просрочено.
> На неделе, отведенной на выполнение задания, у меня была загрузка и некоторые непредвиденные обстоятельства.
> Я решил воспользоваться дополнительным временем из-за проблем с тестирующей системой, чтобы все-таки выполнить
> задание.
> Буду рад, если мое решение будет принято.
> Спасибо за понимание.

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

