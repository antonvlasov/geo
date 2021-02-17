package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

func Save(c *cache, path, name string) (err error) {
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}
	fstring := fmt.Sprintf("%v/%v", strings.TrimSuffix(path, "/"), name)
	var f *os.File
	f, err = os.Create(fstring)
	if err != nil {
		return
	}
	defer f.Close()
	var b []byte
	b, err = json.Marshal(c)
	if err != nil {
		return
	}
	_, err = f.Write(b)
	return
}
func Load(c *cache, path, name string) (err error) {
	var f *os.File
	fullname := fmt.Sprintf("%v/%v", strings.TrimSuffix(path, "/"), name)
	f, err = os.Open(fullname)
	if err != nil {
		return
	}
	defer f.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, f)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &c)
	if err != nil {
		return
	}
	return
}
func (c cache) MarshalJSON() ([]byte, error) {
	res := make([]byte, 0)
	res = append(res, []byte("{\"Fields\":{")...)
	for key := range c.Fields {
		res = append(res, []byte(fmt.Sprintf("\"%v\":", key))...)
		inn, err := json.Marshal(c.Fields[key])
		if err != nil {
			return nil, err
		}
		res = append(res, inn...)
		res = append(res, []byte(",")...)
	}
	res = append(res[:len(res)-1], []byte("},")...)

	res = append(res, []byte("\"Exps\":")...)
	inn, err := json.Marshal(c.Exps)
	if err != nil {
		return nil, err
	}
	res = append(res, inn...)
	res = append(res, []byte("}")...)
	return res, nil
}

func (c *cache) UnmarshalJSON(b []byte) error {
	*c = *(NewCache().(*cache))
	temp := make(map[string]interface{})
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}
	fieldsInt, ok := temp["Fields"]
	if !ok {
		return errors.New("wrong data")
	}
	fields, ok := fieldsInt.(map[string]interface{})
	if !ok {
		return errors.New("wrong data")
	}
	for key := range fields {
		tempb, err := json.Marshal(fields[key])
		if err != nil {
			return err
		}

		var str string
		err = json.Unmarshal(tempb, &str)
		if err == nil {
			c.Fields[key] = str
			continue
		}

		var rlist RList
		err = json.Unmarshal(tempb, &rlist)
		if err == nil {
			c.Fields[key] = rlist
			continue
		}

		var hmap Hashmap
		err = json.Unmarshal(tempb, &hmap)
		if err == nil {
			c.Fields[key] = hmap
			continue
		} else {
			return err
		}
	}
	exps, ok := temp["Exps"]
	if !ok {
		return errors.New("wrong data")
	}

	tempb, err := json.Marshal(exps)
	if err != nil {
		return err
	}

	exp := NewExpirations()
	err = json.Unmarshal(tempb, &exp)
	if err == nil {
		*exp.m = sync.Mutex{}
		c.Exps = exp
	} else {
		return err
	}
	return nil
}
func (l RList) MarshalJSON() ([]byte, error) {
	arr := make([]string, l.Value.Len())
	i := 0
	for iter := l.Value.Front(); iter != nil; iter = iter.Next() {
		arr[i] = iter.Value.(string)
		i++
	}
	bytes, err := json.Marshal(arr)
	if err != nil {
		return bytes, err
	}
	bytes = append([]byte("{\"Value\":"), bytes...)
	bytes = append(bytes, []byte("}")...)
	return bytes, nil
}
func (l *RList) UnmarshalJSON(b []byte) error {
	*l = NewRList()
	b = bytes.TrimLeft(b, "{\"Value\":")
	b = bytes.TrimRight(b, "}")
	var arr []string
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return err
	}
	for i := range arr {
		l.Value.PushBack(arr[i])
	}
	return nil
}
