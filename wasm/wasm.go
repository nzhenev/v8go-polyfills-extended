package wasm

import (
	"github.com/nzhenev/v8go-polyfills-extended/utils"

	v8 "github.com/ionos-cloud/v8go"
)

// Option ...
type Option func(*Module)

// Module ...
type Module struct {
	ModulePath string

	utils.Injector
}

// New ...
func New(opts ...Option) *Module {
	m := new(Module)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithModulePath ...
func WithModulePath(path string) Option {
	return func(m *Module) {
		m.ModulePath = path
	}
}

// Inject ...
func (m *Module) Inject(iso *v8.Isolate, global *v8.ObjectTemplate) error {
	return nil
}
