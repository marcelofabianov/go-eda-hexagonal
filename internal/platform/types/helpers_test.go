package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func mustNewTestUUID(t *testing.T) types.UUID {
	t.Helper()
	id, err := types.NewUUID()
	require.NoError(t, err)
	return id
}
