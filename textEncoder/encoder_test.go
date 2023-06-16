package textEncoder

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

	err := InjectTo(iso, global)
	if err != nil {
		t.Error(err)
	}

	ctx := v8go.NewContext(iso, global)
	if err := console.InjectTo(ctx); err != nil {
		t.Error(err)
	}

	val, err := ctx.RunScript(`const encoder = new TextEncoder()
	console.log(typeof encoder.encode);
	const view = encoder.encode('eyJhdWQiOiJlNjI0YTdjMi0zZTUzLTQ2NTktOGY5Yi1kN2MxOWZjZjAxZjciLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vMjRiMDgwY2QtNTg3NC00NGFiLTk4NjItOGQ3ZTBlMDc4MWFiL3YyLjAiLCJpYXQiOjE2MzkxMzU3NDYsIm5iZiI6MTYzOTEzNTc0NiwiZXhwIjoxNjM5MTM5NjQ2LCJuYW1lIjoiQXNoaXNoIFNoYXJtYSAoRGV2T24pIiwib2lkIjoiOTZmODM2N2QtY2M2NC00NjMwLWI0MGQtYTUwNTVjMjAwOGVkIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiYXNoaXNoLnNoYXJtYUBkZXZvbi5ubCIsInJoIjoiMC5BUUlBellDd0pIUllxMFNZWW8xLURnZUJxOEtuSk9aVFBsbEdqNXZYd1pfUEFmY0NBTzguIiwic3ViIjoiLVNDRE5lR2IwVVc1TzZ5NkoxMERyNWhFZWxIR0lSdU5uNnd3NTZuMHRyMCIsInRpZCI6IjI0YjA4MGNkLTU4NzQtNDRhYi05ODYyLThkN2UwZTA3ODFhYiIsInV0aSI6InJfZERnTWdncGtPN01pQnhPNndTQUEiLCJ2ZXIiOiIyLjAifQ')
	console.log("=>", view); 
	console.log(typeof view); //expecting the type as Object (uint8Array)

	const utf8 = new ArrayBuffer(7);
	let encodedResults = encoder.encodeInto('H€llo', utf8);
	console.log("=>", utf8, encodedResults.read, encodedResults.written); 
	console.log(typeof utf8); //expecting the type as Object (uint8Array)

	utf8 `, "encoder.js")
	if err != nil {
		t.Error(err)
	}

	ok := val.IsArrayBuffer()
	if ok {
		bytes := []byte(val.String())
		fmt.Println("returned val is array", bytes)
	} else {
		fmt.Println("returned val is not array", val.Object().Value)
	}

}
