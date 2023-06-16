package timers

import (
	"fmt"
	"sync"

	"github.com/nzhenev/v8go-polyfills-extended/utils"

	v8 "github.com/nzhenev/v8go"
)

// Timers ...
type Timers struct {
	tt            map[int32]*Timer
	nextTimeoutID int32

	sync.RWMutex
	utils.Injector
}

// New ...
func New() *Timers {
	t := new(Timers)
	t.tt = make(map[int32]*Timer)

	return t
}

// Inject ...
func (t *Timers) Inject(iso *v8.Isolate, global *v8.ObjectTemplate) error {
	time := New()

	for _, f := range []struct {
		Name string
		Func func() v8.FunctionCallback
	}{
		{"setTimeout", time.GetSetTimeoutFunctionCallback},
		{"clearTimeout", time.GetClearTimeoutFunctionCallback},
	} {
		fn := v8.NewFunctionTemplate(iso, f.Func())

		if err := global.Set(f.Name, fn, v8.ReadOnly); err != nil {
			return fmt.Errorf("v8-polyfills/listener: %w", err)
		}
	}

	return nil
}

// GetSetTimeoutFunctionCallback ...
func (t *Timers) GetSetTimeoutFunctionCallback() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		ctx := info.Context()

		id, err := t.startNewTimer(info.This(), info.Args())
		if err != nil {
			v, err := utils.NewInt32Value(ctx, 0)
			if err != nil {
				return nil
			}

			return v
		}

		v, err := utils.NewInt32Value(ctx, id)
		if err != nil {
			return nil
		}

		return v
	}
}

// GetClearTimeoutFunctionCallback ...
func (t *Timers) GetClearTimeoutFunctionCallback() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) > 0 && args[0].IsInt32() {
			t.clear(args[0].Int32())
		}

		return nil
	}
}

func (t *Timers) clear(id int32) {
	t.Lock()
	defer t.Unlock()

	if timer, ok := t.tt[id]; ok {
		timer.Clear()
		delete(t.tt, id)
	}
}

func (t *Timers) startNewTimer(this v8.Valuer, args []*v8.Value) (int32, error) {
	if len(args) <= 0 {
		return 0, fmt.Errorf("1 argument required, but only 0 present")
	}

	fn, err := args[0].AsFunction()
	if err != nil {
		return 0, err
	}
	timeout := args[1].Int32()

	var restArgs []v8.Valuer
	if len(args) > 2 {
		restArgs = make([]v8.Valuer, 0)
		for _, arg := range args[2:] {
			restArgs = append(restArgs, arg)
		}
	}

	t.Lock()
	defer t.Unlock()

	t.nextTimeoutID++
	timer := NewTimer(t.nextTimeoutID, timeout, func() { _, _ = fn.Call(this, restArgs...) })
	t.tt[timer.ID] = timer

	timer.Start()

	return timer.ID, nil
}
