package platformerrors

import (
	"errors"
	"payter-bank/internal/api"
)

var (
	ErrInternal = errors.New("failed to complete request. please try again later")
)

func MakeApiError(code int, message string) *api.ApiError {
	return api.NewErrorWithCode(code, message)
}
