package timers

import (
	"testing"

	v8 "github.com/ionos-cloud/v8go"
	"github.com/stretchr/testify/assert"
)

func TestNewTimer(t *testing.T) {
	tt := New()
	assert.NotNil(t, tt)
}

func Test_SetTimeout(t *testing.T) {
	ctx, err := newContextWithTimers()
	assert.NoError(t, err)

	v, err := ctx.RunScript("setTimeout(() => { return 123 }, 1000)", "timer.js")
	assert.NoError(t, err)
	assert.NotNil(t, v)

	assert.Equal(t, int32(1), v.Int32())
}

func newContextWithTimers() (*v8.Context, error) {
	iso := v8.NewIsolate()
	global := v8.NewObjectTemplate(iso)

	t := New()
	if err := t.Inject(iso, global); err != nil {
		return nil, err
	}

	return v8.NewContext(iso, global), nil
}
