package base64

import (
	"fmt"

	b64 "encoding/base64"

	v8 "github.com/nzhenev/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/utils"
)

// Base64 ...
type Base64 struct {
	utils.Injector
}

// New ...
func New() *Base64 {
	return &Base64{}
}

// GetBtoaFunctionCallback ...
func (b *Base64) GetBtoaFunctionCallback() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			return nil
		}

		s := args[0].String()
		b, err := b64.StdEncoding.DecodeString(s)
		if err != nil {
			return newStringValue(ctx, "")
		}

		return newStringValue(ctx, string(b))
	}
}

// GetAtobFunctionCallback ...
func (b *Base64) GetAtobFunctionCallback() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			return nil
		}

		s := args[0].String()
		b := b64.StdEncoding.EncodeToString([]byte(s))

		return newStringValue(ctx, b)
	}
}

// Inject ...
func (b *Base64) Inject(iso *v8.Isolate, global *v8.ObjectTemplate) error {
	base64 := New()

	for _, f := range []struct {
		Name string
		Func func() v8.FunctionCallback
	}{
		{"btoa", base64.GetBtoaFunctionCallback},
		{"atob", base64.GetAtobFunctionCallback},
	} {
		fn := v8.NewFunctionTemplate(iso, f.Func())

		if err := global.Set(f.Name, fn, v8.ReadOnly); err != nil {
			return fmt.Errorf("v8-polyfills/base64: %w", err)
		}
	}

	return nil
}

func newStringValue(ctx *v8.Context, str string) *v8.Value {
	iso := ctx.Isolate()
	val, _ := v8.NewValue(iso, str)

	return val
}
