package listener

import (
	"fmt"

	v8 "github.com/ionos-cloud/v8go"
)

// Option ...
type Option func(*Listener)

// Listener ...
type Listener struct {
	in  map[string]chan *v8.Object
	out map[string]chan *v8.Value
}

// New ...
func New(opt ...Option) *Listener {
	c := new(Listener)
	c.in = make(map[string]chan *v8.Object)
	c.out = make(map[string]chan *v8.Value)

	for _, o := range opt {
		o(c)
	}

	return c
}

// WithEvents ...
func WithEvents(name string, in chan *v8.Object, out chan *v8.Value) Option {
	return func(l *Listener) {
		l.in[name] = in
		l.out[name] = out
	}
}

// Inject ...
func (l *Listener) Inject(iso *v8.Isolate, global *v8.ObjectTemplate) error {
	ctxFn := v8.NewFunctionTemplate(iso, l.GetFunctionCallback())

	if err := global.Set("addListener", ctxFn, v8.ReadOnly); err != nil {
		return fmt.Errorf("v8-polyfills/listener: %w", err)
	}

	return nil
}

// GetFunctionCallback ...
func (l *Listener) GetFunctionCallback() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		ctx := info.Context()
		args := info.Args()

		if len(args) <= 1 {
			err := fmt.Errorf("addListener: expected 2 arguments, got %d", len(args))

			return newErrorValue(ctx, err)
		}

		fn, err := args[1].AsFunction()
		if err != nil {
			err := fmt.Errorf("%w", err)

			return newErrorValue(ctx, err)
		}

		chn, ok := l.in[args[0].String()]
		if !ok {
			err := fmt.Errorf("addListener: event %s not found", args[0].String())

			return newErrorValue(ctx, err)
		}

		go func(chn chan *v8.Object, fn *v8.Function) {
			for e := range chn {
				v, err := fn.Call(ctx.Global(), e)
				if err != nil {
					fmt.Printf("addListener: %v", err)
				}

				l.out[args[0].String()] <- v
			}
		}(chn, fn)

		return v8.Undefined(ctx.Isolate())
	}
}

func newErrorValue(ctx *v8.Context, err error) *v8.Value {
	iso := ctx.Isolate()
	e, _ := v8.NewValue(iso, fmt.Sprintf("addListener: %v", err))

	return e
}
