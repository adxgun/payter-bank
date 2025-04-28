package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code  int
	Error *ApiError
	Data  interface{}
}

func (r Response) Marshal() ([]byte, error) {
	if r.Error == nil && r.Data == nil {
		return nil, errors.New("no data or error to marshal")
	}

	if r.Error != nil {
		data, err := json.Marshal(ErrorResponse{Error: r.Error.Message})
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	data, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}
	return data, nil

}

type HttpHandler func(ctx *gin.Context) Response

func Wrap(handler HttpHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := handler(ctx)

		data, err := resp.Marshal()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Internal Server error",
			})
			return
		}

		ctx.Data(resp.Code, "application/json", data)
	}
}

type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return e.Message
}

func newError(message string) *ApiError {
	return &ApiError{
		Message: message,
	}
}

func NewErrorWithCode(code int, message string) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
	}
}

func BadRequest(message string) Response {
	return Response{
		Code:  http.StatusBadRequest,
		Error: newError(message),
	}
}

func Error(err error) Response {
	var e *ApiError
	if errors.As(err, &e) {
		return Response{
			Code:  e.Code,
			Error: e,
		}
	}

	return Response{
		Code: http.StatusInternalServerError,
		Error: &ApiError{
			Message: err.Error(),
		},
	}
}

func PreConditionFailed(message string) Response {
	return Response{
		Code:  http.StatusPreconditionFailed,
		Error: newError(message),
	}
}

func Unauthorized(message string) Response {
	return Response{
		Code:  http.StatusUnauthorized,
		Error: newError(message),
	}
}

func Forbidden(message string) Response {
	return Response{
		Code:  http.StatusForbidden,
		Error: newError(message),
	}
}

func NotFound(message string) Response {
	return Response{
		Code:  http.StatusNotFound,
		Error: newError(message),
	}
}

func ServerError(message string) Response {
	return Response{
		Code:  http.StatusInternalServerError,
		Error: newError(message),
	}
}

func OK(message string, data interface{}) Response {
	return Response{
		Code: http.StatusOK,
		Data: SuccessResponse{
			Data:    data,
			Message: message,
		},
	}
}
