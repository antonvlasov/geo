package cache

import (
	"testing"
)

func TestCreateFile(t *testing.T) {
	c := (NewCache()).(*cache)
	resp, err := c.Set([]string{"key", "value"})
	if err != nil {
		t.Error(err)
	}
	if resp != "OK" {
		t.Errorf("expected OK, got %v", resp)
	}
	resp, err = c.HSet([]string{"hashmap", "hash1", "val1", "hash2", "val2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "(integer) 2" {
		t.Errorf("expected (integer) 1, got %v", resp)
	}
	resp, err = c.RPush([]string{"list", "1", "2", "3"})
	if err != nil {
		t.Error(err)
	}
	if resp != "(integer) 3" {
		t.Errorf("expected (integer) 3, got %v", resp)
	}
	resp, err = c.Expire([]string{"key", "2000"})
	if err != nil {
		t.Error(err)
	}
	if resp != "(integer) 1" {
		t.Errorf("expected (integer) 1, got %v", resp)
	}

	err = Save(c, "/home/antonvlasov/heheeheheh/", "ehhe")
	if err != nil {
		t.Error(err)
	}

	cc := NewCache().(*cache)
	err = Load(cc, "/home/antonvlasov/heheeheheh", "ehhe")
	if err != nil {
		t.Error(err)
	}
	go cc.StartCleaner()
	resp, err = cc.Get([]string{"key"})
	if err != nil {
		t.Error(err)
	}
	if resp != "value" {
		t.Errorf("expected %v, got %v", "value", resp)
	}
	resp, err = cc.HGet([]string{"hashmap", "hash1"})
	if err != nil {
		t.Error(err)
	}
	if resp != "val1" {
		t.Errorf("expected %v, got %v", "val1", resp)
	}
	resp, err = cc.HGet([]string{"hashmap", "hash2"})
	if err != nil {
		t.Error(err)
	}
	if resp != "val2" {
		t.Errorf("expected %v, got %v", "val2", resp)
	}
	resp, err = cc.LPop([]string{"list", "0", "-1"})
	if err != nil {
		t.Error(err)
	}
	if resp != "1)1\n2)2\n3)3\n" {
		t.Errorf("expected %v, got %v", "1)1\n2)2\n3)3\n", resp)
	}
}
