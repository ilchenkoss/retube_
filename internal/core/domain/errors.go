package domain

import "errors"

var ErrAlreadyExist = errors.New("already exists")
var ErrNotFound = errors.New("not found")
var ErrUserRecursion = errors.New("user recursion")
