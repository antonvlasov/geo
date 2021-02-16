package cache

import (
	"fmt"
	"testing"
	"time"
)

var keys = []string{
	"firstname",
	"lastname",
	"age",
}
var fields = []interface{}{
	"Anton",
	"Vlasov",
	"20",
}

func TestKeys(t *testing.T) {
	c := (NewCache()).(*cache)

	for i := range keys {
		c.Set([]string{keys[i], fields[i].(string)})
		//c.write(keys[i], fields[i], time.Time{})
	}

	exp := `1) "age"`
	resp, err := c.Keys([]string{`a??`})
	if err != nil {
		t.Error(err)
	}
	if resp != exp {
		t.Errorf("expected:\n %v, got:\n %v", exp, resp)
	}
}
func TestDel(t *testing.T) {
	c := (NewCache()).(*cache)

	for i := range keys {
		c.Set([]string{keys[i], fields[i].(string)})
		//c.write(keys[i], fields[i], time.Time{})
	}

	response, err := c.Del([]string{keys[0], "nonexistant", keys[2]})
	if err != nil {
		t.Error(err)
	}
	if response != "(integer) 2" {
		t.Errorf("expected (integer) 2, got %v", response)
	}
	for key := range c.fields {
		if key != keys[1] || c.fields[key] != fields[1] {
			t.Errorf("expected: %v %v, got %v %v", keys[1], fields[1], key, c.fields[key])
		}
	}

	response, err = c.Del([]string{})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}
}

func TestGet(t *testing.T) {

	c := (NewCache()).(*cache)

	for i := range keys {
		c.Set([]string{keys[i], fields[i].(string)})
	}

	resp, err := c.Get([]string{keys[0], keys[1]})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}
	resp, err = c.Get([]string{})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}

	resp, err = c.Get([]string{keys[0]})
	if err != nil {
		t.Error(err)
	}
	if resp != fields[0] {
		t.Errorf("expected %v, got %v", fields[0], resp)
	}
	resp, err = c.Get([]string{"nonexistant"})
	if resp != "(nil)" {
		t.Errorf("expected (nil), got %v", resp)
	}
}

func TestSet(t *testing.T) {
	c := (NewCache()).(*cache)

	referenceMap := make(map[string]interface{})

	for i := range keys {
		referenceMap[keys[i]] = fields[i]
		resp, err := c.Set([]string{keys[i], fields[i].(string)})
		if err != nil {
			t.Error(err)
		}
		if resp != "OK" {
			t.Errorf("expected OK, got %v", resp)
		}
	}

	referenceMap[keys[2]] = "27"
	resp, err := c.Set([]string{keys[2], "27"})
	if err != nil {
		t.Error(err)
	}
	if resp != "OK" {
		t.Errorf("expected OK, got %v", resp)
	}

	_, err = c.Set([]string{"age"})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}
	for key, value := range referenceMap {
		val := c.read(key)
		if val != value {
			t.Errorf("expected %v, got %v", value, val)
		}
	}
}

func TestHSet(t *testing.T) {
	c := (NewCache()).(*cache)

	for i := range keys {
		_, err := c.HSet([]string{keys[i], fields[i].(string)})
		switch err.(type) {
		case ArgsError:
			break
		default:
			t.Error(err)
		}
	}
	_, err := c.HSet([]string{"key", "hash1", "val1", "hash2"})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}

	k := []string{"hmap1"}
	h := []string{"hash1", "hash2"}
	v := []string{"value1", "value2", "value3"}

	resp, err := c.HSet([]string{k[0], h[0], v[0]})
	if err != nil {
		t.Error(err)
	}
	if resp != "(integer) 1" {
		t.Errorf("expected response (integer) 1 got %v", resp)
	}
	stored := c.read(k[0]).(Hashmap)
	val := stored.value[h[0]]
	if val != v[0] {
		t.Errorf("expected %v, got %v", v[0], val)
	}

	resp, err = c.HSet([]string{k[0], h[1], v[1], h[0], v[2]})
	if err != nil {
		t.Error(err)
	}
	if resp != "(integer) 2" {
		t.Errorf("expected response (integer) 2 got %v", resp)
	}
	stored = c.read(k[0]).(Hashmap)
	val = stored.value[h[0]]
	if val != v[2] {
		t.Errorf("expected %v, got %v", v[2], val)
	}
	val = stored.value[h[1]]
	if val != v[1] {
		t.Errorf("expected %v, got %v", v[1], val)
	}
}

func TestHGet(t *testing.T) {
	c := (NewCache()).(*cache)

	_, err := c.HGet([]string{"key", "hash1", "val1", "hash2"})
	switch err.(type) {
	case ArgsError:
		break
	default:
		t.Error(err)
	}

	c.Set([]string{keys[0], fields[0].(string)})
	_, err = c.HGet([]string{keys[0], fields[0].(string)})
	if err == nil || err.Error() != "Requested field is of type string" {
		t.Error(err)
	}

	k := []string{"hmap1"}
	h := []string{"hash1", "hash2"}
	v := []string{"value1", "value2", "value3"}

	response, err := c.HGet([]string{k[0], h[0]})
	if err != nil {
		t.Error(err)
	}
	if response != "(nil)" {
		t.Errorf("expected (nil), got %v", response)
	}

	c.HSet([]string{k[0], h[0], v[0]})

	response, err = c.HGet([]string{k[0], h[0]})
	if err != nil {
		t.Error(err)
	}
	if response != v[0] {
		t.Errorf("expected %v, got %v", v[0], response)
	}

	c.HSet([]string{k[0], h[1], v[1], h[0], v[2]})

	response, err = c.HGet([]string{k[0], h[0]})
	if response != v[2] {
		t.Errorf("expected %v, got %v", v[2], response)
	}
	response, err = c.HGet([]string{k[0], h[1]})
	if response != v[1] {
		t.Errorf("expected %v, got %v", v[1], response)
	}
}
func TestList(t *testing.T) {
	c := (NewCache()).(*cache)

	lists := []string{"list1", "list2", "list3", "list4"}

	var arr1 []string
	var arr2 []string
	n := 10
	for i := 0; i < n; i += 1 {
		resp, err := c.LPush([]string{lists[0], fmt.Sprint(n - i - 1)})
		if err != nil {
			t.Error(err)
		}
		if resp != fmt.Sprintf("(integer) %v", i+1) {
			t.Errorf("expected resoponse %v, got %v", fmt.Sprintf("(integer) %v", i+1), resp)
		}
		resp, err = c.RPush([]string{lists[2], fmt.Sprint(i)})
		if err != nil {
			t.Error(err)
		}
		if resp != fmt.Sprintf("(integer) %v", i+1) {
			t.Errorf("expected resoponse %v, got %v", fmt.Sprintf("(integer) %v", i+1), resp)
		}
		arr1 = append(arr1, fmt.Sprint(i))
		arr2 = append(arr2, fmt.Sprint(n-i-1))
	}
	resp, err := c.LPush(append([]string{lists[1]}, arr2...))
	if err != nil {
		t.Error(err)
	}
	if resp != fmt.Sprintf("(integer) %v", n) {
		t.Errorf("expected resoponse %v, got %v", fmt.Sprintf("(integer) %v", n), resp)
	}
	resp, err = c.RPush(append([]string{lists[3]}, arr1...))
	if err != nil {
		t.Error(err)
	}
	if resp != fmt.Sprintf("(integer) %v", n) {
		t.Errorf("expected resoponse %v, got %v", fmt.Sprintf("(integer) %v", n), resp)
	}

	for i := 0; i < n; i += 1 {
		var prevResp string
		for j := 0; j < len(lists); j += 1 {
			resp, err := c.LGet([]string{lists[j], fmt.Sprint(i)})
			if err != nil {
				t.Error(err)
			}
			if j > 0 && i > 0 {
				if resp != prevResp {
					t.Errorf("expected %v, got %v", prevResp, resp)
				}
			}
			prevResp = resp

			resp, err = c.LSet([]string{lists[j], fmt.Sprint(i), fmt.Sprint(i * 10)})
			if err != nil {
				t.Error(err)
			}
			if resp != "OK" {
				t.Errorf("expected response %v, got %v", "OK", resp)
			}
		}
	}

	var prevResp string
	for i := 0; i < n; i += 1 {
		for j := 0; j < len(lists)/2; j += 1 {
			resp, err := c.LPop([]string{lists[j]})
			if err != nil {
				t.Error(err)
			}
			if j > 0 && i > 0 {
				if resp != prevResp {
					t.Errorf("expected %v, got %v", prevResp, resp)
				}
			}
			prevResp = resp
		}
	}
	for j := 0; j < len(lists)/2; j += 1 {
		resp, err := c.LPop([]string{lists[j]})
		if err != nil {
			t.Error(err)
		}
		if resp != "(nil)" {
			t.Errorf("expected %v, got %v", "(nil)", resp)
		}
	}

	for i := 0; i < n; i += 1 {
		for j := len(lists) / 2; j < len(lists); j += 1 {
			resp, err := c.RPop([]string{lists[j]})
			if err != nil {
				t.Error(err)
			}
			if j > len(lists)/2 && i > 0 {
				if resp != prevResp {
					t.Errorf("expected %v, got %v", prevResp, resp)
				}
			}
			prevResp = resp
		}
	}
	for j := len(lists) / 2; j < len(lists); j += 1 {
		resp, err := c.RPop([]string{lists[j]})
		if err != nil {
			t.Error(err)
		}
		if resp != "(nil)" {
			t.Errorf("expected %v, got %v", "(nil)", resp)
		}
	}
}
func TestRangePop(t *testing.T) {
	c := (NewCache()).(*cache)

	lists := []string{"list1", "list2"}

	n := 10
	for i := 0; i < n; i += 1 {
		c.LPush([]string{lists[0], fmt.Sprint(n - i - 1)})
		c.RPush([]string{lists[1], fmt.Sprint(i)})
	}

	resp, err := c.LPop([]string{lists[0], "2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)0\n2)1\n" {
		t.Errorf("expected 1)0\n2)1\n got %v", resp)
	}
	resp, err = c.RPop([]string{lists[1], "2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)9\n2)8\n" {
		t.Errorf("expected 1)0\n2)1\n got %v", resp)
	}

	resp, err = c.LPop([]string{lists[0], "0", "-2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)2\n2)3\n3)4\n4)5\n5)6\n6)7\n7)8\n" {
		t.Errorf("expected 1)2\n2)3\n3)4\n4)5\n5)6\n6)7\n7)8\n got %v", resp)
	}
	resp, err = c.RPop([]string{lists[1], "0", "-2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)6\n2)5\n3)4\n4)3\n5)2\n6)1\n7)0\n" {
		t.Errorf("expected 1)6\n2)5\n3)4\n4)3\n5)2\n6)1\n7)0\n got %v", resp)
	}

	resp, err = c.LPop([]string{lists[0], "-1"})
	if err == nil || err.Error() != "count must be positive" {
		t.Error(err)
	}
	resp, err = c.RPop([]string{lists[1], "-1"})
	if err == nil || err.Error() != "count must be positive" {
		t.Error(err)
	}

	resp, err = c.LPop([]string{lists[0], "2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)9\n" {
		t.Errorf("expected 1)9\n got %v", resp)
	}
	resp, err = c.RPop([]string{lists[1], "2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)7\n" {
		t.Errorf("expected 1)9\n got %v", resp)
	}
}
func TestExpire(t *testing.T) {
	c := (NewCache()).(*cache)
	go c.StartCleaner()

	expireIn := 1
	for i := range keys {
		resp, err := c.Set([]string{keys[i], fields[i].(string), "EX", fmt.Sprint(expireIn)})
		if err != nil {
			t.Error(err)
		}
		if resp != "OK" {
			t.Errorf("expected OK, got %v", resp)
		}
	}

	c.HSet([]string{"hashmap", "hash", "value"})
	c.Expire([]string{"hashmap", fmt.Sprint(expireIn)})

	c.LPush([]string{"list", "1", "2", "3"})
	c.Expire([]string{"list", fmt.Sprint(3 * expireIn)})

	for i := range keys {
		resp, err := c.Get([]string{keys[i]})
		if err != nil {
			t.Error(err)
		}
		if resp != fields[i].(string) {
			t.Errorf("expected %v, got %v", fields[i].(string), resp)
		}
	}

	resp, err := c.HGet([]string{"hashmap", "hash"})
	if err != nil {
		t.Error(err)
	}
	if resp != "value" {
		t.Errorf("expected %v got %v", "value", resp)
	}

	resp, err = c.LGet([]string{"list", "0"})
	if err != nil {
		t.Error(err)
	}
	if resp != "3" {
		t.Errorf("expected %v got %v", "3", resp)
	}

	time.Sleep(time.Duration(2 * expireIn * int(time.Second)))

	for i := range keys {
		resp, err := c.Get([]string{keys[i]})
		if err != nil {
			t.Error(err)
		}
		if resp != "(nil)" {
			t.Errorf("expected %v got %v", "(nil)", resp)
		}
	}
	resp, err = c.HGet([]string{"hashmap", "hash"})
	if err != nil {
		t.Error(err)
	}
	if resp != "(nil)" {
		t.Errorf("expected %v got %v", "(nil)", resp)
	}

	resp, err = c.LGet([]string{"list", "0"})
	if err != nil {
		t.Error(err)
	}
	if resp != "3" {
		t.Errorf("expected %v got %v", "3", resp)
	}

	time.Sleep(time.Duration(2 * int(time.Second)))

	resp, err = c.LPop([]string{"list", "0", "-1"})
	if err != nil {
		t.Error(err)
	}
	if resp != "(nil)" {
		t.Errorf("expected %v got %v", "(nil)", resp)
	}
}