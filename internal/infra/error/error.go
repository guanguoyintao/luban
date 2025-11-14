// Package error 错误处理框架
package error

import (
	"fmt"
	"runtime"
)

// ErrorCode 错误代码类型
type ErrorCode string

// 预定义错误代码
const (
	CodeInternalError     ErrorCode = "INTERNAL_ERROR"
	CodeInvalidParameter ErrorCode = "INVALID_PARAMETER"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeTimeout          ErrorCode = "TIMEOUT"
	CodeRateLimited      ErrorCode = "RATE_LIMITED"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// Error 自定义错误类型
type Error struct {
	Code       ErrorCode
	Message    string
	Details    map[string]interface{}
	StackTrace []string
	Cause      error
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Details:    make(map[string]interface{}),
		StackTrace: captureStackTrace(),
	}
}

// Newf 创建格式化的错误
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Details:    make(map[string]interface{}),
		StackTrace: captureStackTrace(),
		Cause:      err,
	}
}

// Wrapf 包装现有错误并格式化消息
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *Error {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现错误解包
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithDetail 添加错误详情
func (e *Error) WithDetail(key string, value interface{}) *Error {
	e.Details[key] = value
	return e
}

// WithDetails 批量添加错误详情
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// GetCode 获取错误代码
func (e *Error) GetCode() ErrorCode {
	return e.Code
}

// GetMessage 获取错误消息
func (e *Error) GetMessage() string {
	return e.Message
}

// GetDetails 获取错误详情
func (e *Error) GetDetails() map[string]interface{} {
	return e.Details
}

// GetStackTrace 获取堆栈跟踪
func (e *Error) GetStackTrace() []string {
	return e.StackTrace
}

// captureStackTrace 捕获堆栈跟踪
func captureStackTrace() []string {
	const maxDepth = 32
	pc := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pc) // 跳过前3个调用帧
	
	frames := runtime.CallersFrames(pc[:n])
	var stack []string
	
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	
	return stack
}

// Is 检查错误是否为指定类型
func Is(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// As 将错误转换为指定类型
func As(err error) (*Error, bool) {
	if e, ok := err.(*Error); ok {
		return e, true
	}
	return nil, false
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	Handle(err error) error
	CanHandle(err error) bool
}

// BaseErrorHandler 基础错误处理器
type BaseErrorHandler struct {
	code ErrorCode
}

// NewBaseErrorHandler 创建基础错误处理器
func NewBaseErrorHandler(code ErrorCode) *BaseErrorHandler {
	return &BaseErrorHandler{code: code}
}

// CanHandle 检查是否能处理错误
func (h *BaseErrorHandler) CanHandle(err error) bool {
	if e, ok := As(err); ok {
		return e.Code == h.code
	}
	return false
}

// ErrorChain 错误处理链
type ErrorChain struct {
	handlers []ErrorHandler
}

// NewErrorChain 创建错误处理链
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		handlers: []ErrorHandler{},
	}
}

// AddHandler 添加处理器
func (c *ErrorChain) AddHandler(handler ErrorHandler) {
	c.handlers = append(c.handlers, handler)
}

// Handle 处理错误
func (c *ErrorChain) Handle(err error) error {
	for _, handler := range c.handlers {
		if handler.CanHandle(err) {
			if handledErr := handler.Handle(err); handledErr != nil {
				return handledErr
			}
		}
	}
	return err
}

// HTTPError HTTP错误响应
type HTTPError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToHTTPError 转换为HTTP错误
func ToHTTPError(err error) HTTPError {
	if e, ok := As(err); ok {
		return HTTPError{
			Code:    getHTTPStatusCode(e.Code),
			Message: e.Message,
			Details: e.Details,
		}
	}
	
	return HTTPError{
		Code:    500,
		Message: "Internal Server Error",
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// getHTTPStatusCode 获取HTTP状态码
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case CodeInvalidParameter:
		return 400
	case CodeNotFound:
		return 404
	case CodeAlreadyExists:
		return 409
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeTimeout:
		return 408
	case CodeRateLimited:
		return 429
	case CodeServiceUnavailable:
		return 503
	default:
		return 500
	}
}