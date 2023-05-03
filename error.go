package anya

import (
	"errors"
	"fmt"
)

var (
	errBodyNil = errors.New("web: can't read nil body")
)

func errInvalidPath(path string) error {
	return errors.New(fmt.Sprintf("web: invalid path: %v", path))
}

func errPathRegistered(path string) error {
	return errors.New(fmt.Sprintf("web: path already registered: %v", path))
}

func errNodeConflict(exist string, current string) error {
	return errors.New(fmt.Sprintf("web: exists %v node conflicts with current %v node", exist, current))
}

func errRouteNotExist(path string) error {
	return errors.New(fmt.Sprintf("web: path: %v doesn't exist", path))
}

func errKeyNotExist(key string) error {
	return errors.New(fmt.Sprintf("web: key: %s not exist", key))
}
