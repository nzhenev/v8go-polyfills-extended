package console

import (
	"os"
	"testing"

	"github.com/esoptra/v8go"
)

func TestInject(t *testing.T) {
	t.Parallel()

	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	if err := InjectTo(ctx, WithOutput(os.Stdout)); err != nil {
		t.Error(err)
	}

	if _, err := ctx.RunScript("console.log(1111)", ""); err != nil {
		t.Error(err)
	}
}
