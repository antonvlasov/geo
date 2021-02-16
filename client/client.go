package client

import (
	"fmt"
	"net"
	"strings"
)

func Keys(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("KEYS %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func Del(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("DEL %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func Get(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("GET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func Set(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("SET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func HGet(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("HGET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func HSet(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("HSET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func LPush(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("LPUSH %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func RPush(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("RPUSH %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func LPop(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("LPOP %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func RPop(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("RPOP %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func LGet(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("LGET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
func LSet(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(fmt.Sprintf("LSET %v\r\n", strings.Trim(fmt.Sprint(args), "[]"))))
	return err
}
