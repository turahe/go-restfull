package response

import "net/http"

type Envelope struct {
	Code       int     `json:"code"`
	Message    string  `json:"message"`
	Data       any     `json:"data"`
	NextCursor *string `json:"next_cursor,omitempty"`
	PrevCursor *string `json:"prev_cursor,omitempty"`
	Error      any     `json:"error"`
}

func JSON(c Context, httpStatus int, code int, message string, data any, err any) {
	c.JSON(httpStatus, Envelope{
		Code:    code,
		Message: message,
		Data:    data,
		Error:   err,
	})
}

func JSONCursor(c Context, httpStatus int, code int, message string, data any, nextCursor *string, prevCursor *string) {
	c.JSON(httpStatus, Envelope{
		Code:       code,
		Message:    message,
		Data:       data,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Error:      nil,
	})
}

func OK(c Context, code int, message string, data any) {
	JSON(c, http.StatusOK, code, message, data, nil)
}

func OKCursor(c Context, code int, message string, data any, nextCursor *string, prevCursor *string) {
	JSONCursor(c, http.StatusOK, code, message, data, nextCursor, prevCursor)
}

func Created(c Context, code int, message string, data any) {
	JSON(c, http.StatusCreated, code, message, data, nil)
}

func BadRequest(c Context, code int, message string, err any) {
	JSON(c, http.StatusBadRequest, code, message, nil, err)
}

func Unauthorized(c Context, code int, message string, err any) {
	JSON(c, http.StatusUnauthorized, code, message, nil, err)
}

func Forbidden(c Context, code int, message string, err any) {
	JSON(c, http.StatusForbidden, code, message, nil, err)
}

func NotFound(c Context, code int, message string, err any) {
	JSON(c, http.StatusNotFound, code, message, nil, err)
}

func InternalServerError(c Context, code int, message string, err any) {
	JSON(c, http.StatusInternalServerError, code, message, nil, err)
}

// Context is a tiny interface to avoid coupling pkg/response to Gin.
type Context interface {
	JSON(code int, obj any)
}

