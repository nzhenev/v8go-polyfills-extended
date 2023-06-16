/*
 * Copyright (c) 2021 Xingwang Liao
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

package fetch

import (
	_ "embed"
	"fmt"

	"github.com/nzhenev/v8go"
)

//go:embed internal/headers.js
var headers string

//go:embed internal/response.js
var response string

//go:embed internal/body.js
var body string

//go:embed internal/request.js
var request string

func InjectTo(iso *v8go.Isolate, global *v8go.ObjectTemplate, opt ...Option) error {
	f := NewFetcher(opt...)

	fetchFn := v8go.NewFunctionTemplate(iso, f.GetFetchFunctionCallback())

	if err := global.Set("fetch", fetchFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch: %w", err)
	}
	RequestFn := v8go.NewFunctionTemplate(iso, RequestCallbackFunc)

	if err := global.Set("Request", RequestFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch RequestFn: %w", err)
	}
	return nil
}

func InjectWithFetcherTo(iso *v8go.Isolate, global *v8go.ObjectTemplate, f Fetcher) error {
	fetchFn := v8go.NewFunctionTemplate(iso, f.GetFetchFunctionCallback())
	if err := global.Set("fetch", fetchFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch: %w", err)
	}

	RequestFn := v8go.NewFunctionTemplate(iso, RequestCallbackFunc)

	if err := global.Set("Request", RequestFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch RequestFn: %w", err)
	}

	return nil
}

func InjectWithCtx(ctx *v8go.Context, opt ...Option) error {
	f := NewFetcher(opt...)

	iso := ctx.Isolate()

	fetchFn := v8go.NewFunctionTemplate(iso, f.GetFetchFunctionCallback())

	con := v8go.NewObjectTemplate(iso)

	if err := con.Set("fetch", fetchFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch: %w", err)
	}
	RequestFn := v8go.NewFunctionTemplate(iso, RequestCallbackFunc)
	if err := con.Set("Request", RequestFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/fetch RequestFn: %w", err)
	}
	return nil
}

func InjectHTTPProperties(ctx *v8go.Context) error {
	_, err := ctx.RunScript(headers, "headers.js")
	if err != nil {
		return fmt.Errorf("v8go-polyfills/headers inject: %w", err)
	}

	_, err = ctx.RunScript(response, "response.js")
	if err != nil {
		return fmt.Errorf("v8go-polyfills/response inject: %w", err)
	}

	// _, err = ctx.RunScript(body, "body.js")
	// if err != nil {
	// 	return fmt.Errorf("v8go-polyfills/body inject: %w", err)
	// }

	// _, err = ctx.RunScript(request, "request.js")
	// if err != nil {
	// 	return fmt.Errorf("v8go-polyfills/request inject: %w", err)
	// }
	return nil
}
