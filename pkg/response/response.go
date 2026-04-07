package response

import "net/http"

// Envelope is the canonical API response format.
// Tests (and handler clients) unmarshal into this type.
type Envelope struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    any     `json:"data"`
	Next    *bool   `json:"next,omitempty"`
	Prev    *bool   `json:"prev,omitempty"`
	// Cursor-specific fields are optional and used by some pagination styles.
	NextCursor *string `json:"nextCursor,omitempty"`
	PrevCursor *string `json:"prevCursor,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func JSON(c Context, httpStatus int, code int, message string, data any, err any) {
	c.JSON(httpStatus, Envelope{
		Code:    code,
		Message: message,
		Data:    data,
		Error:   err,
	})
}

func JSONPaginated(c Context, httpStatus int, code int, message string, data any, next bool, prev bool) {
	c.JSON(httpStatus, Envelope{
		Code:    code,
		Message: message,
		Data:    data,
		Next:    &next,
		Prev:    &prev,
		Error:   nil,
	})
}

func OK(c Context, code int, message string, data any) {
	JSON(c, http.StatusOK, code, message, data, nil)
}

func OKPaginated(c Context, code int, message string, data any, next bool, prev bool) {
	JSONPaginated(c, http.StatusOK, code, message, data, next, prev)
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

func Conflict(c Context, code int, message string, err any) {
	JSON(c, http.StatusConflict, code, message, nil, err)
}

func InternalServerError(c Context, code int, message string, err any) {
	JSON(c, http.StatusInternalServerError, code, message, nil, err)
}

// Context is a tiny interface to avoid coupling pkg/response to Gin.
type Context interface {
	JSON(code int, obj any)
}
