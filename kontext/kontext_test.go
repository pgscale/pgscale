package kontext

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	e := New()

	t.Run("Get-Set", func(t *testing.T) {
		e.Set("mykey", "myvalue")
		value := e.Get("mykey")
		require.NotNil(t, value)
		require.IsType(t, "", value)
	})

	t.Run("Clone", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			e.Set(fmt.Sprintf("mykey-%d", i), fmt.Sprintf("myvalue-%d", i))
		}

		cloned := e.Clone()
		require.True(t, reflect.DeepEqual(e, cloned))
	})
}
