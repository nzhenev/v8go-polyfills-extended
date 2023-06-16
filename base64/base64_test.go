package base64

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v8 "github.com/nzhenev/v8go"
)

func TestAtob(t *testing.T) {
	ctx, err := newV8goContext()
	assert.NoError(t, err)
	defer ctx.Close()

	v, err := ctx.RunScript("atob('test')", "test.js")
	assert.NoError(t, err)

	assert.Equal(t, "dGVzdA==", v.String())
}

func TestBtob(t *testing.T) {
	ctx, err := newV8goContext()
	assert.NoError(t, err)
	defer ctx.Close()

	v, err := ctx.RunScript("btoa('dGVzdA==')", "test.js")
	assert.NoError(t, err)

	assert.Equal(t, "test", v.String())
}

func newV8goContext() (*v8.Context, error) {
	iso := v8.NewIsolate()
	global := v8.NewObjectTemplate(iso)

	b := New()

	if err := b.Inject(iso, global); err != nil {
		return nil, err
	}

	return v8.NewContext(iso, global), nil
}
