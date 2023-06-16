package base64

import (
	stdBase64 "encoding/base64"

	"github.com/esoptra/v8go"
)

type Base64 interface {
	GetAtobFunctionCallback() v8go.FunctionCallback
	GetBtoaFunctionCallback() v8go.FunctionCallback
}

type base64 struct {
}

func NewBase64() Base64 {
	return &base64{}
}

/*
https://developer.mozilla.org/en-US/docs/Web/API/WindowOrWorkerGlobalScope/atob
*/
func (b *base64) GetAtobFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			// TODO: v8go can't throw a error now, so we return an empty string
			return newStringValue(ctx, "")
		}

		encoded := args[0].String()
		if encoded == "" {
			return newStringValue(ctx, "")
		}
		byts, err := stdBase64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return newStringValue(ctx, "")
		}
		return newStringFromByteArray(ctx, byts)
	}
}

/*
https://developer.mozilla.org/en-US/docs/Web/API/WindowOrWorkerGlobalScope/btoa
*/
func (b *base64) GetBtoaFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			return newStringValue(ctx, "")
		}

		str := args[0].String()
		encoded := stdBase64.RawStdEncoding.EncodeToString([]byte(str))
		return newStringValue(ctx, encoded)
	}
}

func newStringValue(ctx *v8go.Context, str string) *v8go.Value {
	iso := ctx.Isolate()
	val, _ := v8go.NewValue(iso, str)
	return val
}

func newStringFromByteArray(ctx *v8go.Context, byts []byte) *v8go.Value {
	iso := ctx.Isolate()
	val, _ := v8go.NewStringFromByteArray(iso, byts)
	return val
}
