package ptr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dyleme/Notifier/pkg/utils/ptr"
)

func TestZeroIfNil(t *testing.T) {
	t.Parallel()

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		i := ptr.On(1)

		actual := ptr.ZeroIfNil(i)
		assert.Equal(t, 1, actual)
	})

	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		var i *int

		actual := ptr.ZeroIfNil(i)
		assert.Equal(t, 0, actual)
	})
}
