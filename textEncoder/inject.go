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

	"github.com/nzhenev/v8go"
)

// uses encode func to create encodeinto wrapper method
func InjectWith(iso *v8go.Isolate, global *v8go.ObjectTemplate, opt ...Option) (*v8go.Context, error) {
	e := NewEncode(opt...)
	encodeFnTmp := v8go.NewFunctionTemplate(iso, e.TextEncoderFunctionCallback())
	if err := global.Set("TextEncoder1", encodeFnTmp); err != nil {
		return nil, fmt.Errorf("v8go-polyfills/textEncoder global.set: %w", err)
	}

	ctx := v8go.NewContext(iso, global)

	_, err := ctx.RunScript(`class TextEncoder {
             encode(usvstring){
                     const encoder = new TextEncoder1();
                     return encoder.encode(usvstring);
             }
             encodeInto(usvstring, utf8) { 
                     const encoder = new TextEncoder1()
                     console.log(typeof encoder.encode);
                     let res = encoder.encode(usvstring)
                     for (var i =0; i < res.length; i++) {
                             utf8[i]=res[i]
                     }
                     return {read: usvstring.length, written: utf8.length }; 
             }
     }`, "encoder.js")
	if err != nil {
		return nil, fmt.Errorf("v8go-polyfills/textEncoder Ra(): %w", err)
	}
	return ctx, nil
}

func InjectTo(iso *v8go.Isolate, global *v8go.ObjectTemplate, opt ...Option) error {
	e := NewEncode(opt...)
	encodeFnTmp := v8go.NewFunctionTemplate(iso, e.TextEncoderFunctionCallback())
	if err := global.Set("TextEncoder", encodeFnTmp); err != nil {
		return fmt.Errorf("v8go-polyfills/textEncoder global.set: %w", err)
	}
	return nil
}
