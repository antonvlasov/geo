# Имплементация im-memory Redis кэша
- Язык реализации go
- Возможность хранить строки, списки и словари
- Реализована возможность установить TTL на любой ключ
- Работа через telnet при помощи REST API
- Запуск через docker-compose up
- Возможность сохранения на диск
- Масштабирование
- Результат покрытия тестами в файле coverage.out
# Инструкция для запуска
Необходимо иметь: git;docker; docker-compose.
 - Скачать содежримое репозитория командой ```git clone https://github.com/antonvlasov/geo```
 - Перейти в скачанную директорию : ```cd geo```
 - Запустить контейнер командой ```docker-compose up```
# Инструкция по использованию
 1) Подключиться к запущенному серверу по telnet к порту 7089. Например, при запуске на локальной машине команда ```telnet localhost 7089``` в командной строке.
2) Использовать команды для взаимодействия с сервером. Например ``` SET name Anton EX 60```.
# Список команд
### KEYS pattern
Возвращает список ключей, удовлетворяющих glob-style образцу.
Пример: 
```
SET name Anton
OK
SET age 20
OK
KEYS *
1) "name"
2) "age"
```
### GET key
Возвращает значение по ключу key. Если по ключу нет значения, возвращается nil. Если значение по ключу не типа строка, возвращается ошибка.
Пример:
```
SET name Anton
OK
GET name
Anton
GET PUE
(nil)
```
### SET key value [EX seconds]
Устанавливает значение по ключу key равным строке value. Если указан дополнительный параметр EX seconds, значение очищается через заданное количество секунд. Значение seconds 0 означает хранение без ограничений по времени.
Пример:
```
SET key1 value1 EX 20
OK
```
### EXPIRE key seconds
Устанавливает время жизни значения по ключу. Если по указаному ключу существует значение, возвращает 1, иначе 0.
Пример:
```
SET key1 1     
OK
EXPIRE key3 30
(integer) 0
EXPIRE key1 20
(integer) 1
...after some time
GET key1
(nil)
```
### HSET key field value [field value ...]
Устанавливает поле словаря field, являющимся значением по ключу key равным value. Возможно установить сразу несколько полей. Если значение указанного ключа является другим типом, возвращается ошибка. Возвращает количество затронутых полей.
Пример:
```
SET key1 1
OK
HSET key1 hash1 val1
Requested field is of type string
HSET key1 hash1 val1 hash2 val2
(integer) 2
```
### HGET key field
Возвращает значение поля field словаря по ключу key. Если по ключу key находится другой тип, возвращается ошибка. Если значение поля field не задано или значение по ключу key не задано, возвращается nil.
Пример:
```
SET name Anton
OK
HSET key1 hash1 val1 hash2 val2
(integer) 2

HGET key1 hash1
val1
HGET key1 hash2 
val2
HGET key1 hash3
(nil)
HGET key434 hash2
(nil)
HGET name hash
Requested field is of type string
```
### LPUSH key element [element ...]
Вставляет элементы слева в список по ключу key. Если элементов несколько, они вставляются так, как будто для каждого из них по порядку была бы вызвана эта команда. Если значения по ключу не существовало, список создается. Если по ключу значение другого типа, возвращается ошибка.
Пример:
```
SET key1 val1 
OK
LPUSH key1 1
Requested field is of type string
LPUSH list1 1 2 3
(integer) 3
LPOP list1 0 -1
1)3
2)2
3)1
```
### RPUSH key element [element ...]
Вставляет элементы справа в список по ключу key. Если элементов несколько, они вставляются так, как будто для каждого из них по порядку была бы вызвана эта команда. Если значения по ключу не существовало, список создается. Если по ключу значение другого типа, возвращается ошибка.
Пример:
```
SET key1 val1 
OK
RPUSH key1 1
Requested field is of type string
RPUSH list1 1 2 3
(integer) 3
LPOP list1 0 -1
1)1
2)2
3)3
```
### LPOP key [count]
Удаляет и возвращает элемент слева списка. Параметр count - количество удаляемых элементов, может быть либо единственным числом - тогда это количество элементов с края, либо двумя числами - тогда это индексы первого и последнего удаляемых элементов. Индексы могут быть отрицательными для доступа с конца списка. Если количество удаляемых элементо превышает количество элементов в списке, возвращается доступное количество. Если по указаному ключу данные другого типа, возвращается ошибка.
Пример:
```
RPUSH list1 1 2 3 4 5 6 7 8 9 10
(integer) 10
LPOP 2
(nil)
LPOP list1 2
1)1
2)2

LPOP list1 2 -2
1)5
2)6
3)7
4)8
5)9

LPOP list1
3
```
### RPOP key [count]
Удаляет и возвращает элемент слева списка. Параметр count - количество удаляемых элементов, может быть либо единственным числом - тогда это количество элементов с края, либо двумя числами - тогда это индексы первого и последнего удаляемых элементов. Индексы могут быть отрицательными для доступа с конца списка. Если количество удаляемых элементо превышает количество элементов в списке, возвращается доступное количество. Если по указаному ключу данные другого типа, возвращается ошибка.
Пример:
```
RPUSH list1 1 2 3 4 5 6 7 8 9 10
(integer) 10
RPOP list1 2
1)10
2)9

RPOP list1 2 -2
1)7
2)6
3)5
4)4
5)3

RPOP list1
1
```
### LSET key index element
Устанавливает значение элемента с индексом index списка по ключу key равным element. Если элемента с этим индексом не существует, возвращается ошибка.
Пример:
```
RPUSH list1 0 1 2 3 4 5 6 7 8 9
(integer) 10
LSET list1 3 30
OK
LGET list1 3
30
LSET list1 20 2
Index out of range
```
### LGET key index
Получает значение элемента с индексом index из списка по ключу key. Если элемента с этим индексом не существует, возвращается ошибка.
Пример:
```
RPUSH list1 0 1 2 3 4 5 6 7 8 9
(integer) 10
LSET list1 3 30
OK
LGET list1 3
30
LSET list1 20 2
Index out of range
```
# Клиент
Все методы доступны в виде функций с подписью вида
```func(conn net.Conn, args []string) error {}```
и описаны в файле client.go
Пример использования:
```
clientConn, err := net.DialTimeout("tcp", "localhost:7089", 0)
	if err != nil {
		t.Error(err)
	}

	err = Set(clientConn, []string{"name", "Anton"})
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(clientConn)
	bytes, err := reader.ReadBytes('\n')
	fmt.Println(string(bytes))
```	
# Сохранение
Для сохранения необходимо использовать команду ```SAVE savename```, где savename - имя сохранения. После этой команды сервер сохранит данные под указанным именем. Для загрузки данных испрользуется команда ```LOAD savename```. В результате этой команды все текущие данные заменяются на данные из сохранения.
# Тесты
Релизовано покрытия тестами более 70% кода и нагрузочные тесты операция записи и чтения, нагрузочный тест операции чтения с использованием нескольких потоков.
### Операция записи:
BenchmarkSet             1000000               173 ns/op
### Операция чтения:
BenchmarkGet             1000000                56.4 ns/op
### Конкурентное чтение с разным количеством потоков:
BenchmarkGetConcurrent            10000000               289 ns/op
BenchmarkGetConcurrent-2          10000000               287 ns/op
BenchmarkGetConcurrent-4          10000000               226 ns/op
BenchmarkGetConcurrent-8          10000000               238 ns/op 