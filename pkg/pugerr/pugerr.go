package pugerr

import "fmt"

const (
	Successful   = 0x0000 // 请求成功
	ApiViolation = 0x7531 // API请求参数校验不通过（如: NotBlank NotEmpty）
	SystemBusy   = -1     // 系统繁忙，请稍候再试
	Undefined    = 0xFFFF
)

type Error interface {
	error

	ErrorCode() int

	Message() string

	ErrCause() error
}

type AppError struct {
	errorCode int
	message   string
	err       error
}

func (ths AppError) ErrorCode() int {
	return ths.errorCode
}

func (ths AppError) Message() string {
	return ths.message
}

func (ths AppError) ErrCause() error {
	return ths.err
}

func (ths AppError) Error() string {
	if nil != ths.ErrCause() {
		return fmt.Sprintf("Error code: %d, message: %s errCause %s", ths.errorCode, ths.message, ths.ErrCause().Error())
	} else {
		return fmt.Sprintf("Error code: %d, message: %s", ths.errorCode, ths.message)
	}
}

func ViolationError(message string) Error {
	return &AppError{
		ApiViolation,
		message,
		nil,
	}
}

func ViolationErrorWithErr(message string, err error) Error {
	return &AppError{
		ApiViolation,
		message,
		err,
	}
}

func UndefinedError(err error) Error {
	return &AppError{
		Undefined,
		err.Error(),
		err,
	}
}

func SystemBusyError() Error {
	return &AppError{
		SystemBusy,
		"The system is busy, please try again later.",
		nil,
	}
}
