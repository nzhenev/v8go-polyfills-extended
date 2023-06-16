package listener_test

import (
	"testing"

	"github.com/nzhenev/v8go-polyfills-extended/console"
	"github.com/nzhenev/v8go-polyfills-extended/listener"

	"github.com/stretchr/testify/assert"

	v8 "github.com/esoptra/v8go"
)

func BenchmarkEventListenerCall(b *testing.B) {
	iso := v8.NewIsolate()
	global := v8.NewObjectTemplate(iso)

	in := make(chan *v8.Object)
	out := make(chan *v8.Value)

	l := listener.New()
	err := l.Inject(iso, global)
	assert.NoError(b, err)

	ctx := v8.NewContext(iso, global)

	if err := console.InjectTo(ctx); err != nil {
		panic(err)
	}

	_, err = ctx.RunScript("addListener('auth', event => { return event.sourceIP === '127.0.0.1' })", "listener.js")
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		obj, err := newContextObject(ctx)
		assert.NoError(b, err)
		in <- obj

		v := <-out

		assert.NotNil(b, v)
		assert.True(b, v.IsBoolean())
	}
}

func newContextObject(ctx *v8.Context) (*v8.Object, error) {
	iso := ctx.Isolate()
	obj := v8.NewObjectTemplate(iso)

	resObj, err := obj.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range []struct {
		Key string
		Val interface{}
	}{
		{Key: "sourceIP", Val: "127.0.0.1"},
	} {
		if err := resObj.Set(v.Key, v.Val); err != nil {
			return nil, err
		}
	}

	return resObj, nil
}
