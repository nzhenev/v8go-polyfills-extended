package wasm

import (
	"testing"

	v8 "github.com/ionos-cloud/v8go"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New()
	assert.NotNil(t, m)
}

func TestInject(t *testing.T) {
	ctx, err := newV8goContext()
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
}

func newV8goContext() (*v8.Context, error) {
	iso := v8.NewIsolate()
	global := v8.NewObjectTemplate(iso)

	m := New(WithModulePath("./test.wasm"))

	if err := m.Inject(iso, global); err != nil {
		return nil, err
	}

	return v8.NewContext(iso, global), nil
}
