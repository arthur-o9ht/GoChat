package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type UserInfo struct {
	UserName string
	PassWd   string
}

func regUser(regUser UserInfo) (bool, error) {
	c, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return false, err
	}
	defer c.Close()
	if _, err := c.Do("hmset", redis.Args{}.Add(userListKey+regUser.UserName).AddFlat(&regUser)...); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func checkUser(regUser UserInfo) (map[string]string, error) {
	c, err := redis.Dial("tcp", "localhost:6379")
	errStr := make(map[string]string)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return errStr, err
	}
	defer c.Close()

	if res, err := redis.StringMap(c.Do("hgetall", userListKey+regUser.UserName)); err != nil {
		return errStr, err
	} else {
		return res, nil
	}
}
