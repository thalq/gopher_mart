package errors

import "errors"

var TooManyRequestsError = errors.New("too many requests")
var InternalServerError = errors.New("internal server error")
