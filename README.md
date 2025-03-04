## Описание проекта

Приложение для управления задачами, позволяющее создавать, изменять, удалять задачи и осуществлять их поиск. Также предусмотрена функция аутентификации пользователей.

## Список выполненных задач со звёздочкой

- Настройка порта веб-сервера, пути к файлу базы данных и пароля через переменные окружения.
- Гибкая система повторений задач:

    - Еженедельное выполнение задач в выбранные дни недели.
    - Ежемесячное выполнение задач по заданным числам.

- Функция поиска задач по заголовку, комментариям и дате.
- Возможность аутентификации при наличии установленного пароля.

## Инструкция по запуску кода

### Запуск
Чтобы запустить проект локально, выполните следующую команду в терминале, находясь в каталоге проекта:

    go run .

### Определение переменных окружения
В проекте поддерживается настройка трёх переменных окружения:

    PORT: задаёт порт для веб-сервера.
    TODO_DBFILE: указывает путь к файлу базы данных.
    TODO_PASSWORD: определяет пароль для последующей аутентификации.

Эти переменные можно определить в файле .env, расположенном в корневой директории проекта. Пример структуры файла:

    TODO_PASSWORD = "roottoor"
    PORT = "7054"
    TODO_DBFILE = "./database/scheduler.db"

### Проект в браузере
Чтобы получить доступ к сервису после его локального запуска, в браузере необходимо перейти по адресу localhost:<номер порта>. По умолчанию порт установлен на значение 7540.

## Запуск тестов
Для запуска тестирования всего приложения используйте команду:

    go test ./tests

Некоторые параметры можно настроить в файле tests/settings.go:
   
    - Port — изменить порт по умолчанию.
    - DBFile — задать путь к файлу базы данных.
    - FullNextDate — установите true для проверки всех условий повторения задач или false для ограниченного набора тестов.
    - Search — установите true для тестирования поиска или false, чтобы отключить его проверку.
    - Token — JWT-токен для аутентификации.
