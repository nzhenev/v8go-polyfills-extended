package utils

import (
	v8 "github.com/esoptra/v8go"
)

// Injector ...
type Injector interface {
	Inject(*v8.Isolate, *v8.ObjectTemplate) error
}
