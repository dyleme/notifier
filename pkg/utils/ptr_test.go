package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Dyleme/Notifier/pkg/utils"
)

func TestZeroIfNil(t *testing.T) {
	t.Parallel()

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		i := utils.Ptr(1)

		actual := utils.ZeroIfNil(i)
		assert.Equal(t, 1, actual)
	})

	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		var i *int

		actual := utils.ZeroIfNil(i)
		assert.Equal(t, 0, actual)
	})
}
