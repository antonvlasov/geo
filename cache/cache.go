package cache

import (
	"container/heap"
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
)

const savepath string = "./saves"

type ArgsError struct {
	msg string
}

func (err ArgsError) Error() string {
	return err.msg
}

type Hashmap struct {
	Hashmap map[string]string
}

func NewHashmap() Hashmap {
	return Hashmap{make(map[string]string)}
}

func (h *Hashmap) Read(key string) (value string, ok bool) {
	value, ok = h.Hashmap[key]
	return
}
func (h *Hashmap) Write(key, value string) {
	h.Hashmap[key] = value
}

type RList struct {
	Value *list.List
}

func NewRList() RList {
	return RList{list.New()}
}

type Cache interface {
	//must lock objects recieved from read
	read(string) interface{}
	write(string, interface{})
	delete(string) bool
	setExpiration(key string, expires time.Duration)

	StartCleaner()

	HandleRequest(method string, args []string) (string, error)
	Keys(args []string) (response string, err error)
	Del(args []string) (response string, err error)
	Get(args []string) (response string, err error)
	Set(args []string) (response string, err error)
	HSet(args []string) (response string, err error)
	HGet(args []string) (response string, err error)
	LPush(args []string) (response string, err error)
	RPush(args []string) (response string, err error)
	LPop(args []string) (response string, err error)
	RPop(args []string) (response string, err error)
	LSet(args []string) (response string, err error)
	LGet(args []string) (response string, err error)
	Expire(args []string) (response string, err error)
	Save(args []string) (response string, err error)
	Load(args []string) (response string, err error)
}

func NewCache() Cache {
	return &cache{Fields: make(map[string]interface{}), m: &sync.RWMutex{}, Exps: NewExpirations()}
}

type Expirations struct {
	Expirations []expiration
	Indexes     map[string]int
	m           *sync.Mutex
}

func NewExpirations() Expirations {
	return Expirations{make([]expiration, 0), make(map[string]int), &sync.Mutex{}}
}
func (exp Expirations) Len() int {
	return len(exp.Expirations)
}
func (exp Expirations) Less(i, j int) bool {
	return exp.Expirations[i].Expires.Before(exp.Expirations[j].Expires)
}
func (exp Expirations) Swap(i, j int) {
	exp.Expirations[i], exp.Expirations[j] = exp.Expirations[j], exp.Expirations[i]
	exp.Indexes[exp.Expirations[i].Field] = i
	exp.Indexes[exp.Expirations[j].Field] = j
}
func (exp *Expirations) Pop() interface{} {
	n := len(exp.Expirations)
	res := exp.Expirations[n-1]
	delete(exp.Indexes, res.Field)
	exp.Expirations = exp.Expirations[:n-1]
	return res
}
func (exp *Expirations) Push(val interface{}) {
	record := val.(expiration)
	exp.Expirations = append(exp.Expirations, record)
	exp.Indexes[record.Field] = len(exp.Expirations) - 1
}

type expiration struct {
	Field   string
	Expires time.Time
}
type cache struct {
	Fields map[string]interface{}
	m      *sync.RWMutex
	Exps   Expirations
}

// Mutex must be rlocked before calling read
func (c *cache) read(key string) interface{} {
	return c.Fields[key]
}

// Mutex must be locked before calling write
func (c *cache) write(key string, val interface{}) {
	c.Fields[key] = val
}

// Mutex must be locked before calling delete
func (c *cache) delete(key string) bool {
	_, ok := c.Fields[key]
	if ok {
		delete(c.Fields, key)
	}
	return ok
}

// c.exp.m must be locked when calling this function
func (c *cache) setExpiration(key string, expires time.Duration) {
	if expires != 0 {
		timeStamp := time.Now().Add(expires)
		heap.Push(&c.Exps, expiration{key, timeStamp})
	} else {
		index, ok := c.Exps.Indexes[key]
		if ok {
			heap.Remove(&c.Exps, index)
		}
	}
}

// Starts cleaning the expired fields, function blocks. Uses c.m, watch for deadlock
func (c *cache) StartCleaner() {
	for {
		time.Sleep(50 * time.Millisecond)
		c.Exps.m.Lock()
		if c.Exps.Len() != 0 {
			// closest expiration
			expiration := c.Exps.Expirations[0]
			for ; expiration.Expires.Before(time.Now()); expiration = c.Exps.Expirations[0] {
				if expiration.Expires.Before(time.Now()) {
					heap.Pop(&c.Exps)
					c.m.Lock()
					c.delete(expiration.Field)
					c.m.Unlock()
					if c.Exps.Len() == 0 {
						break
					}
				}
			}
		}
		c.Exps.m.Unlock()
	}
}
func (c *cache) Keys(args []string) (response string, err error) {
	if len(args) != 1 {
		err = ArgsError{"Expected format: KEYS pattern"}
		return
	}
	glob := glob.MustCompile(args[0])
	var keys []string
	i := 1
	c.m.RLock()
	for key := range c.Fields {
		match := glob.Match(key)
		if match {
			keys = append(keys, fmt.Sprintf("%v) \"%v\"", i, key))
			i += 1
		}
	}
	c.m.RUnlock()
	response = strings.Join(keys, "\n")
	return
}
func (c *cache) Del(args []string) (response string, err error) {
	if len(args) == 0 {
		err = ArgsError{"Expected format: DEL key [key ...]"}
		return
	}
	counter := 0
	c.m.Lock()
	for i := range args {
		deleted := c.delete(args[i])
		if deleted {
			counter += 1
		}
	}
	c.m.Unlock()
	response = fmt.Sprintf("(integer) %v", counter)
	return
}
func (c *cache) Get(args []string) (response string, err error) {
	if len(args) != 1 {
		err = ArgsError{"Expected format: GET key"}
		return
	}
	c.m.RLock()
	defer c.m.RUnlock()
	value := c.read(args[0])
	switch value.(type) {
	case string:
		response = value.(string)
	case nil:
		response = "(nil)"
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", value))
	}
	return
}
func (c *cache) Set(args []string) (response string, err error) {
	if len(args) != 2 && len(args) != 4 {
		err = ArgsError{"Expected format: SET key value [EX seconds]"}
		return
	}

	c.Exps.m.Lock()
	c.m.Lock()

	c.write(args[0], args[1])
	if len(args) == 4 {
		var exp time.Duration
		var secs int
		if args[2] == "EX" {
			secs, err = strconv.Atoi(args[3])
			if err != nil {
				return
			}

			exp = time.Duration(secs * int(time.Second) / int(time.Nanosecond))
			if exp != 0 {
				c.setExpiration(args[0], exp) // Setting the correct expiration
			}
		} else {
			err = ArgsError{"Expected format: METHOD [EX seconds] key [arguments]"}
			return
		}
	}
	c.m.Unlock()
	c.Exps.m.Unlock()

	response = "OK"
	return
}
func (c *cache) Expire(args []string) (response string, err error) {
	if len(args) != 2 {
		err = ArgsError{"Expected format: EXPIRE key seconds"}
		return
	}
	c.m.RLock()
	stored := c.read(args[0])
	c.m.RUnlock()
	if stored == nil {
		response = "(integer) 0"
	} else {
		var exp time.Duration
		var secs int
		secs, err = strconv.Atoi(args[1])
		if err != nil {
			return
		}
		exp = time.Duration(secs * int(time.Second) / int(time.Nanosecond))
		c.Exps.m.Lock()
		c.setExpiration(args[0], exp)
		c.Exps.m.Unlock()
		response = "(integer) 1"
	}
	return
}
func (c *cache) HSet(args []string) (response string, err error) {
	n := len(args)
	if n < 3 || n%2 == 0 {
		err = ArgsError{"Expected format: HSET key field value [field value ...]"}
		return
	}
	counter := 0
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	for i := 1; i < n; i += 2 {
		switch stored.(type) {
		case nil:
			emptymap := NewHashmap()
			c.write(args[0], emptymap)
			stored = c.read(args[0])
		case Hashmap:
			break
		default:
			err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
			return
		}
	}
	for i := 1; i < n; i += 2 {
		value := stored.(Hashmap)
		value.Hashmap[args[i]] = args[i+1]
		counter += 1
	}

	response = fmt.Sprintf("(integer) %v", counter)
	return
}
func (c *cache) HGet(args []string) (response string, err error) {
	if len(args) != 2 {
		err = ArgsError{"Expected format: HGET key field"}
		return
	}
	c.m.RLock()
	defer c.m.RUnlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case Hashmap:
		hmap := stored.(Hashmap)
		val, ok := hmap.Read(args[1])
		if !ok {
			response = "(nil)"
		} else {
			response = val
		}
	case nil:
		response = "(nil)"
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	return
}
func (c *cache) LPush(args []string) (response string, err error) {
	if len(args) < 2 {
		err = ArgsError{"Expected format: LPUSH key element [element ...]"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case nil:
		list := NewRList()
		c.write(args[0], list)
		stored = c.read(args[0])
	case RList:
		break
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	list := stored.(RList)
	for i := 1; i < len(args); i += 1 {
		list.Value.PushFront(args[i])
	}

	response = fmt.Sprintf("(integer) %v", list.Value.Len())
	return
}
func (c *cache) RPush(args []string) (response string, err error) {
	if len(args) < 2 {
		err = ArgsError{"Expected format: RPUSH key element [element ...]"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case nil:
		list := NewRList()
		c.write(args[0], list)
		stored = c.read(args[0])
	case RList:
		break
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	list := stored.(RList)
	for i := 1; i < len(args); i += 1 {
		list.Value.PushBack(args[i])
	}

	response = fmt.Sprintf("(integer) %v", list.Value.Len())
	return
}
func (c *cache) LPop(args []string) (response string, err error) {
	if len(args) < 1 || len(args) > 3 {
		err = ArgsError{"Expected format: LPOP key [count]"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case nil:
		response = "(nil)"
	case RList:
		list := stored.(RList)
		if len(args) > 1 {
			response, err = list.poprange(args, "LPOP")
			if err != nil {
				return
			}
		} else {
			response = list.Value.Remove(list.Value.Front()).(string)
		}
		if list.Value.Len() == 0 {
			c.delete(args[0])
		}
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	return
}
func (c *cache) RPop(args []string) (response string, err error) {
	if len(args) < 1 || len(args) > 3 {
		err = ArgsError{"Expected format: RPOP key [count]"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case nil:
		response = "(nil)"
	case RList:
		list := stored.(RList)
		if len(args) > 1 {
			response, err = list.poprange(args, "RPOP")
			if err != nil {
				return
			}
		} else {
			response = list.Value.Remove(list.Value.Front()).(string)
		}
		if list.Value.Len() == 0 {
			c.delete(args[0])
		}
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	return
}
func (l *RList) poprange(args []string, method string) (response string, err error) {
	var popped strings.Builder
	var start int
	var end int
	if len(args) == 3 {
		start, err = strconv.Atoi(args[1])
		if err != nil {
			return
		}
		end, err = strconv.Atoi(args[2])
		if err != nil {
			return
		}
		start, err = formatIndex(start, l.Value.Len())
		if err != nil {
			return
		}
		end, err = formatIndex(end, l.Value.Len())
		if err != nil {
			return
		}
	} else if len(args) == 2 {
		if method == "LPOP" {
			start = 0
			end, err = strconv.Atoi(args[1])
			if err != nil {
				return
			}
			end -= 1
			if end < 0 {
				err = errors.New("count must be positive")
				return
			}
			if end >= l.Value.Len() {
				end = l.Value.Len() - 1
			}
		} else if method == "RPOP" {
			var temp int
			temp, err = strconv.Atoi(args[1])
			if err != nil {
				return
			}
			start = l.Value.Len() - temp
			if start >= l.Value.Len() {
				err = errors.New("count must be positive")
				return
			}
			if start < 0 {
				start = 0
			}
			end = l.Value.Len() - 1
		} else {
			err = errors.New("unknown method")
			return
		}
	}
	if start > end {
		err = errors.New("First index must be less than second")
		return
	}

	if method == "LPOP" {
		iter := l.get(start)
		if iter == nil {
			err = errors.New("Index out of range")
			return
		}
		for i := 0; i <= end-start; i += 1 {
			next := iter.Next()
			element := l.Value.Remove(iter).(string)
			addElement(&popped, &element, i+1)
			iter = next
		}
	} else if method == "RPOP" {
		iter := l.get(end)
		if iter == nil {
			err = errors.New("Index out of range")
			return
		}
		for i := 0; i <= end-start; i += 1 {
			next := iter.Prev()
			element := l.Value.Remove(iter).(string)
			addElement(&popped, &element, i+1)
			iter = next
		}
	} else {
		err = errors.New("unknown method")
		return
	}
	response = popped.String()
	return
}
func addElement(builder *strings.Builder, str *string, i int) error {
	_, err := builder.WriteString(fmt.Sprintf("%v)", i))
	if err != nil {
		return err
	}
	_, err = builder.WriteString(*str)
	if err != nil {
		return err
	}
	_, err = builder.WriteString("\n")
	if err != nil {
		return err
	}
	return nil
}
func (l *RList) get(index int) *list.Element {
	if index < 0 || index > l.Value.Len() {
		return nil
	}
	var iter *list.Element
	if index < l.Value.Len()-index {
		iter = l.Value.Front()
		for count := 0; count < index; count += 1 {
			iter = iter.Next()
		}
	} else {
		iter = l.Value.Back()
		for count := 0; count < l.Value.Len()-index-1; count += 1 {
			iter = iter.Prev()
		}
	}
	return iter
}
func formatIndex(index, length int) (int, error) {
	if index < 0 {
		index = length + index
	}
	if index < 0 {
		index = 0
	}
	if index >= length {
		index = length - 1
	}
	return index, nil
}
func (c *cache) LSet(args []string) (response string, err error) {
	if len(args) != 3 {
		err = ArgsError{"Expected format: LSET key index element"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case RList:
		value := stored.(RList)
		var index int
		index, err = strconv.Atoi(args[1])
		if err != nil {
			return
		}
		elem := value.get(index)
		if elem == nil {
			err = errors.New("Index out of range")
			return
		}

		elem.Value = args[2]
		response = "OK"
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	return
}
func (c *cache) LGet(args []string) (response string, err error) {
	if len(args) != 2 {
		err = ArgsError{"Expected format: LGET key index"}
		return
	}
	c.m.RLock()
	defer c.m.RUnlock()
	stored := c.read(args[0])
	switch stored.(type) {
	case RList:
		value := stored.(RList)
		var index int
		index, err = strconv.Atoi(args[1])
		if err != nil {
			return
		}

		elem := value.get(index)
		if elem == nil {
			err = errors.New("Index out of range")
		}

		response = elem.Value.(string)
	default:
		err = errors.New(fmt.Sprintf("Requested field is of type %T", stored))
		return
	}
	return
}
func (c *cache) Save(args []string) (response string, err error) {
	if len(args) != 1 {
		err = ArgsError{"Expected format: SAVE name"}
		return
	}
	c.m.RLock()
	defer c.m.RUnlock()
	err = Save(c, savepath, args[0])
	if err != nil {
		return
	}
	response = "OK"
	return
}
func (c *cache) Load(args []string) (response string, err error) {
	if len(args) != 1 {
		err = ArgsError{"Expected format: LOAD name"}
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	err = Load(c, savepath, args[0])
	if err != nil {
		return
	}
	response = "OK"
	return
}
func (c *cache) HandleRequest(method string, args []string) (response string, err error) {
	switch method {
	case "KEYS":
		return c.Keys(args)
	case "DEL":
		return c.Del(args)
	case "GET":
		return c.Get(args)
	case "SET":
		return c.Set(args)
	case "HGET":
		return c.HGet(args)
	case "HSET":
		return c.HSet(args)
	case "LPUSH":
		return c.LPush(args)
	case "RPUSH":
		return c.RPush(args)
	case "LPOP":
		return c.LPop(args)
	case "RPOP":
		return c.RPop(args)
	case "LGET":
		return c.LGet(args)
	case "LSET":
		return c.LSet(args)
	case "EXPIRE":
		return c.Expire(args)
	case "SAVE":
		return c.Save(args)
	case "LOAD":
		return c.Load(args)
	default:
		err = errors.New("method does not exist")
		return
	}
}
