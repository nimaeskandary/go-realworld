package util_test

import (
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/util"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

func TestOptionValidation(t *testing.T) {
	v := util.NewValidator()

	t.Run("mo.Option validations", func(t *testing.T) {
		t.Run("treats none as nil", func(t *testing.T) {
			input := struct {
				Value mo.Option[string] `validate:"required"`
			}{
				Value: mo.None[string](),
			}

			err := v.Struct(input)

			assert.Error(t, err)
		})

		t.Run("unwraps option when running validations", func(t *testing.T) {
			// empty string fails "required" validation
			input := struct {
				Value mo.Option[string] `validate:"required"`
			}{
				Value: mo.Some(""),
			}

			err := v.Struct(input)

			assert.Error(t, err)

			// non empty string passes "required" validation
			input = struct {
				Value mo.Option[string] `validate:"required"`
			}{
				Value: mo.Some("foo"),
			}

			err = v.Struct(input)

			assert.NoError(t, err)
		})
	})
}
