package utils

import (
	"testing"

	"github.com/esoptra/v8go"
	"github.com/stretchr/testify/assert"
)

func TestNewInt32Value(t *testing.T) {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	v, err := NewInt32Value(ctx, 123)
	assert.NoError(t, err)

	assert.Equal(t, int32(123), v.Int32())
}
