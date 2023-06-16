package textDecoder

import (
	"fmt"
	"testing"

	"github.com/esoptra/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/console"
)

func TestInject(t *testing.T) {
	t.Parallel()

	iso := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(iso)

	if err := InjectWith(iso, global); err != nil {
		t.Error(err)
	}

	ctx := v8go.NewContext(iso, global)
	if err := console.InjectTo(ctx); err != nil {
		t.Error(err)
	}

	val, err := ctx.RunScript(`const decoder = new TextDecoder()
	const utf8 = new Uint8Array([72,226,130,172,108,108,111,32,87,111,114,108,100]);
	const view = decoder.decode(utf8)
	console.log("=>", view); 

	const decoder1 = new TextDecoder('windows-1251', 'fatal=true')
	let bytes = new Uint8Array([207, 240, 232, 226, 229, 242, 44, 32, 236, 232, 240, 33, 226, 130, 172]);
	const view1 = decoder1.decode(bytes)
	console.log("=>", view1);
	console.log(typeof view); //expecting the type as string
	view `, "encoder.js")
	if err != nil {
		t.Error(err)
	}

	ok := val.IsUint8Array()
	if ok {
		fmt.Println("returned val is array", val.Object().Value)
	} else {
		fmt.Println("returned val is not array", val.Object().Value)
	}
}
