package utils

import "github.com/nzhenev/v8go"

// NewInt32Value ...
func NewInt32Value(ctx *v8go.Context, i int32) (*v8go.Value, error) {
	iso := ctx.Isolate()
	v, err := v8go.NewValue(iso, i)
	if err != nil {
		return nil, err
	}

	return v, nil
}
