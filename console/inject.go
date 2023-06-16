package console

import (
	"errors"
	"fmt"

	"github.com/esoptra/v8go"
)

/*
*
Inject basic console.log support.
*/
func InjectTo(ctx *v8go.Context, opt ...Option) error {
	if ctx == nil {
		return errors.New("v8go-polyfills/console: ctx is required")
	}

	iso := ctx.Isolate()

	c := NewConsole(opt...)

	con := v8go.NewObjectTemplate(iso)

	logFn := v8go.NewFunctionTemplate(iso, c.GetLogFunctionCallback())

	if err := con.Set("log", logFn, v8go.ReadOnly); err != nil {
		return fmt.Errorf("v8go-polyfills/console: %w", err)
	}

	conObj, err := con.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("v8go-polyfills/console: %w", err)
	}

	global := ctx.Global()

	if err := global.Set("console", conObj); err != nil {
		return fmt.Errorf("v8go-polyfills/console: %w", err)
	}

	return nil
}
