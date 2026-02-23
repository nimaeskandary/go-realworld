package util

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/samber/mo"
)

func NewValidator(opts ...validator.Option) *validator.Validate {
	_opts := append(opts, validator.WithRequiredStructEnabled())
	v := validator.New(_opts...)

	RegisterMoOptionType(
		v,
		// can add to these defaults as needed
		mo.Option[string]{},
		mo.Option[int]{},
		mo.Option[int32]{},
		mo.Option[int64]{},
		mo.Option[float32]{},
		mo.Option[float64]{},
		mo.Option[bool]{},
		mo.Option[any]{},
	)

	return v
}

// RegisterMoOptionType handles validating mo.Option[T] types, by unwrapping the some value if present, or returning nil if not present.
// Note, a validation like "required", will fail on None, because it unwraps to nil.
func RegisterMoOptionType(v *validator.Validate, types ...any) {
	v.RegisterCustomTypeFunc(
		func(field reflect.Value) any {
			// safeguard that these methods are avialable and the field is indeed a mo.Option
			isPresentMethod := field.MethodByName("IsPresent")
			mustGetMethod := field.MethodByName("MustGet")

			if isPresentMethod.IsValid() && mustGetMethod.IsValid() {
				isPresent := isPresentMethod.Call(nil)[0].Bool()
				if isPresent {
					return mustGetMethod.Call(nil)[0].Interface()
				}
				return nil
			}
			return nil
		},
		types...,
	)
}
