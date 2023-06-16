package timers

import (
	"fmt"

	"github.com/nzhenev/v8go"
)

func InjectTo(iso *v8go.Isolate, global *v8go.ObjectTemplate) error {
	t := NewTimers()

	for _, f := range []struct {
		Name string
		Func func() v8go.FunctionCallback
	}{
		{Name: "setTimeout", Func: t.GetSetTimeoutFunctionCallback},
		{Name: "setInterval", Func: t.GetSetIntervalFunctionCallback},
		{Name: "clearTimeout", Func: t.GetClearTimeoutFunctionCallback},
		{Name: "clearInterval", Func: t.GetClearIntervalFunctionCallback},
	} {
		fn := v8go.NewFunctionTemplate(iso, f.Func())

		if err := global.Set(f.Name, fn, v8go.ReadOnly); err != nil {
			return fmt.Errorf("v8go-polyfills/timers: %w", err)
		}
	}

	return nil
}
