Итоговый проект
ToDo list - разработка бэкнда web-приложения для планирования задач на день

Выполнены все задания со сзвездочкой

Хендленры находятся в файле todolist.go, функции для работы с бд в файле tasks.go, функции для определения следующей даты по правилам в файле nextdate.go

Добавлено получение переменных среды согласно заданию: TODO_PORT - порт, TODO_DBFILE - путь к файлу БД, TODO_PASSWORD - пароль аутентификации, TODO_WEBDIR - путь к файлам фронтэнда

Переменные среды можно указать в командной строке bash с помошью export, например export TODO_PORT="7540"

Если переменные среды не заполнены:
-Порт по умолчанию: 7540, 
-Путь к файлу БД - папка, из которой запущен исполняемый файл + scheduler.db (если запускать командой go run, то файл создастся во временной папке и при каждом запуске в новой)
-Пароль проверяться не будет
-путь к файлам фронтэнда - ../../web

Настройки для тестирования заполняются в файле tests/settings.go, там можно указать порт в переменной Port и путь файлу БД в переменной DBFile.
Для того чтобы работали тесты нужно заполнить переменную окружения TODO_DBFILE путем к файлу или заполнить путь в файле tests/settings.go, иначе тесты не будут правильно работать - создается файл в разных папках и тесты могут работать неадекватно. Так же для тестов нужно запустить сервер (либо командой "go run ./...", либо запустив исполняемый файл todolist)
Запусить все тесты можно командой "go test -run ^TestDone$ ./tests".

jwt токен формируется по ключу [TODO_PASSWORD + "final"], если заполнить эту переменную среды, то для тестирования нужно заполнить токен, сформированный по такому ключу, в файле settings.go в строке [var Token = ``], например для пароля 123 токен будет eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.jIIppLyFuRhhR5EIXzckZ7ILph20q-aiXpC0NW1Dh8k

Dockerfile собирает контейнер со скомпилированным исполняемым файлом todolist (компилировал командой go build todolist.go nextdate.go tasks.go) и файлами фронтенда в папке web.
В докер файле установленны переменные среды: порт установлен (TODO_PORT="7450"), путь к файлу БД - scheduler.db в текущей папке (TODO_DBFILE="./scheduler.db"), пароль 123 (TODO_PASSWORD="123"), путь к файлам фронтенда (TODO_WEBDIR="./web")
Образ можно собрать командой bash "docker build -t todolist -f build/Dockerfile ." из папки проекта.
Запустить образ можно командой bash "docker run -d -p 7540:7540 todolist". из папки проекта.