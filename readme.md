# Реализация кодогенератора, который ищет методы структуры, помеченные спец меткой и генерирует для них следующий код:
* http-обёртки для этих методов
* проверку авторизации
* проверки метода (GET/POST)
* валидацию параметров
* заполнение структуры с параметрами метода
* обработку неизвестных ошибок
 
 Пример задания взят с coursera golang (см. репозиторий https://github.com/legion15q/Coursera_golang)
Кодогенератор умеет обрабатывать следующие типы полей структуры:
* int
* string

Доступны следующие метки валидатора-заполнятора `apivalidator`:
* required - поле не должно быть пустым (не должно иметь значение по-умолчанию)
* paramname - если указано - то брать из параметра с этим именем, иначе lowercase от имени
* enum - "одно из"
* default - если указано и приходит пустое значение (значение по-умолчанию) - устанавливать то что написано указано в default
* min - >= X для типа int, для строк len(str) >=
* max - <= X для типа int


Авторизация проверяется просто на то что в header пришло значение `100500`

## Структура директории:
* handlers_gen/codegen.go - исходный код генератора
* api.go - передается в кодогенератор.
* main.go - 
* main_test.go - тестирование

##Запуск тестов:
``` shell
# находясь в этой папке
# собирает кодогенератор и сразу же запускает генерацию http-хендлеров для файла api.go, записывая результат в api_handlers.go
go build handlers_gen/* && ./codegen.exe api.go api_handlers.go
# запуск тестов
go test -v
```