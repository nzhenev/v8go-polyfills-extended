package main

import (
	"fmt"
	"time"

	"github.com/nzhenev/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/timers"
)

func main() {
	iso := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(iso)

	if err := timers.InjectTo(iso, global); err != nil {
		panic(err)
	}

	ctx := v8go.NewContext(iso, global)

	val, err := ctx.RunScript(
		"new Promise((resolve) => setTimeout(function(name) {resolve(`Hello, ${name}!`)}, 1000, 'Tom'))",
		"resolve.js",
	)
	if err != nil {
		panic(err)
	}

	proms, err := val.AsPromise()
	if err != nil {
		panic(err)
	}

	done := make(chan bool, 1)

	go func() {
		for proms.State() == v8go.Pending {
			continue
		}

		done <- proms.State() == v8go.Fulfilled
	}()

	select {
	case succ := <-done:
		if !succ {
			panic("except success but not")
		}

		fmt.Println(proms.Result().String())
	case <-time.After(time.Second * 2):
		panic("timeout")
	}
}
