package response

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type captureCtx struct {
	status int
	obj    any
}

func (c *captureCtx) JSON(code int, obj any) {
	c.status = code
	c.obj = obj
}

func TestOK_WritesEnvelope(t *testing.T) {
	t.Parallel()

	ctx := &captureCtx{}
	OK(ctx, 2000001, "ok", map[string]any{"x": 1})

	assert.Equal(t, http.StatusOK, ctx.status)
	env, ok := ctx.obj.(Envelope)
	assert.True(t, ok)
	assert.Equal(t, 2000001, env.Code)
	assert.Equal(t, "ok", env.Message)
	assert.NotNil(t, env.Data)
	assert.Nil(t, env.Error)
}

func TestBadRequest_SetsError(t *testing.T) {
	t.Parallel()

	ctx := &captureCtx{}
	BadRequest(ctx, 4000011, "bad", "oops")

	assert.Equal(t, http.StatusBadRequest, ctx.status)
	env, ok := ctx.obj.(Envelope)
	assert.True(t, ok)
	assert.Equal(t, "bad", env.Message)
	assert.Nil(t, env.Data)
	assert.Equal(t, "oops", env.Error)
}

