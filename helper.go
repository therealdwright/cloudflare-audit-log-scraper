package main

import (
	"errors"
	"os"
	"strconv"
)

var ErrEnvVarEmpty = errors.New("get env: environment variable empty")

// helper functions to handle user input
func getEnvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func getEnvInt(key string) (int, error) {
	s, err := getEnvStr(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}
