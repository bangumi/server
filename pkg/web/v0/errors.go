package v0

import "errors"

var ErrQuery = errors.New("search query is not valid")
var ErrMissingParam = errors.New("search query is missing")
