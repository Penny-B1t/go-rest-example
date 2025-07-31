package error

import "net/http"

type AppError struct {
	StatusCode int
	Code string
	Message string
	Err error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func NewNotFoundError(message string, err error) *AppError {
	if message == "" {
		message = "요청한 리소스를 찾을 수 없습니다."
	}
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message: message,
		Err: err,
	}
}

func NewBadValidateError(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message : "잘못된 요청입니다.",
		Err     : err,
	}
		
}

func NewBadRequestError(message string, err error) *AppError {
	if message == "" {
		message = "잘못된 요청입니다."
	}
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Err: err,
	}
}

func NewInternalServerError(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message: "서버 내부 오류가 발생하였습니다.",
		Err: err,
	}
}