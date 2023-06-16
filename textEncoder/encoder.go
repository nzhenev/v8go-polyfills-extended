/*
 * Copyright (c) 2021 Twintag
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package textEncoder

import (
	"fmt"

	"github.com/esoptra/v8go"
)

type Encoder struct {
}

func NewEncode(opt ...Option) *Encoder {
	c := &Encoder{}

	for _, o := range opt {
		o.apply(c)
	}

	return c
}

// implements pollyfill -> https://developer.mozilla.org/en-US/docs/Web/API/TextEncoder
func (c *Encoder) TextEncoderFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		iso := ctx.Isolate()

		//https://developer.mozilla.org/en-US/docs/Web/API/TextEncoder/encode
		encodeFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			args := info.Args()
			if len(args) <= 0 {
				strErr, _ := v8go.NewValue(iso, "Expected an arguments\n")
				return iso.ThrowException(strErr)
			}
			s := args[0].String()
			v, err := v8go.NewValue(iso, []byte(s))
			if err != nil {
				strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error creating new val: %#v", err))
				return iso.ThrowException(strErr)
			}
			return v
		})

		//https://developer.mozilla.org/en-US/docs/Web/API/TextEncoder/encodeInto
		encodeIntoFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			args := info.Args()
			if len(args) <= 0 {
				strErr, _ := v8go.NewValue(iso, "Expected an arguments\n")

				return iso.ThrowException(strErr)
			}
			s := args[0].String()
			if !args[1].IsArrayBuffer() {
				strErr, _ := v8go.NewValue(iso, "Expected second argument format as ArrayBuffer\n")

				return iso.ThrowException(strErr)
			}

			// outArray := []byte(args[1].String())
			result := make([]uint8, len(s)*3)
			i := 0
			for ; i < len(s); i++ {
				fmt.Printf("%d ", s[i])
				result[i] = s[i]
			}
			// outArray.PutBytes(result[:i])
			obj := info.Context().Global() // create object
			obj.Set("read", int32(i))      // set some properties
			obj.Set("written", int32(len(result[:i])))
			return obj.Value
		})

		resTmp := v8go.NewObjectTemplate(iso)

		if err := resTmp.Set("encode", encodeFnTmp, v8go.ReadOnly); err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error setting encode function template: %#v", err))

			return iso.ThrowException(strErr)
		}

		if err := resTmp.Set("encodeInto", encodeIntoFnTmp, v8go.ReadOnly); err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error setting encodeInto function template: %#v", err))

			return iso.ThrowException(strErr)
		}

		resObj, err := resTmp.NewInstance(ctx)
		if err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error new instance from ctx: %#v", err))

			return iso.ThrowException(strErr)
		}
		return resObj.Value
	}
}
