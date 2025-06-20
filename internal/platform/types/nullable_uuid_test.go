package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNullableUUID_IsValid(t *testing.T) {
	t.Run("valid NullableUUID should return true", func(t *testing.T) {
		testID := mustNewTestUUID(t)
		nu := types.NewValidNullableUUID(testID)
		assert.True(t, nu.IsValid())
	})

	t.Run("null NullableUUID should return false", func(t *testing.T) {
		nu := types.NewNullUUID()
		assert.False(t, nu.IsValid())
	})

	t.Run("NullableUUID with Nil UUID but Valid true should return false", func(t *testing.T) {
		nu := types.NullableUUID{NullUUID: uuid.NullUUID{UUID: uuid.Nil, Valid: true}}
		assert.False(t, nu.IsValid(), "IsValid should be false if UUID is Nil even if Valid is true")
	})

	t.Run("NullableUUID with non-Nil UUID but Valid false should return false", func(t *testing.T) {
		testID := mustNewTestUUID(t)
		nu := types.NullableUUID{NullUUID: uuid.NullUUID{UUID: uuid.UUID(testID), Valid: false}}
		assert.False(t, nu.IsValid(), "IsValid should be false if Valid is false even if UUID is non-Nil")
	})
}

func TestNewNullableUUID(t *testing.T) {
	t.Run("create valid NullableUUID", func(t *testing.T) {
		testID := mustNewTestUUID(t)
		nu := types.NewNullableUUID(testID, true)
		assert.True(t, nu.Valid)
		assert.Equal(t, testID, types.UUID(nu.UUID))
	})

	t.Run("create invalid (null) NullableUUID", func(t *testing.T) {
		testID := mustNewTestUUID(t)
		nu := types.NewNullableUUID(testID, false)
		assert.False(t, nu.Valid)
		retrievedID, ok := nu.GetUUID()
		assert.False(t, ok)
		assert.True(t, retrievedID.IsNil())
	})

	t.Run("create invalid (null) NullableUUID with Nil id", func(t *testing.T) {
		nu := types.NewNullableUUID(types.Nil, false)
		assert.False(t, nu.Valid)
		assert.Equal(t, uuid.Nil, nu.UUID)
	})
}

func TestNewValidNullableUUID(t *testing.T) {
	testID := mustNewTestUUID(t)
	nu := types.NewValidNullableUUID(testID)
	assert.True(t, nu.Valid)
	assert.Equal(t, testID, types.UUID(nu.UUID))
}

func TestNewNullUUID(t *testing.T) {
	nu := types.NewNullUUID()
	assert.False(t, nu.Valid)
	assert.Equal(t, uuid.Nil, nu.UUID)
	id, valid := nu.GetUUID()
	assert.False(t, valid)
	assert.True(t, id.IsNil())
}

func TestNullableUUID_GetUUID(t *testing.T) {
	t.Run("get from valid NullableUUID", func(t *testing.T) {
		testID := mustNewTestUUID(t)
		nu := types.NewValidNullableUUID(testID)
		retrievedID, ok := nu.GetUUID()
		assert.True(t, ok)
		assert.Equal(t, testID, retrievedID)
	})

	t.Run("get from invalid (null) NullableUUID", func(t *testing.T) {
		nu := types.NewNullUUID()
		retrievedID, ok := nu.GetUUID()
		assert.False(t, ok)
		assert.Equal(t, types.Nil, retrievedID)
	})
}

func TestNullableUUID_JSONEncoding(t *testing.T) {
	validTestID := mustNewTestUUID(t)
	validNu := types.NewValidNullableUUID(validTestID)
	invalidNu := types.NewNullUUID()

	t.Run("MarshalJSON for valid NullableUUID", func(t *testing.T) {
		jsonData, err := json.Marshal(validNu)
		require.NoError(t, err)
		expectedJSON := fmt.Sprintf("\"%s\"", validTestID.String())
		assert.JSONEq(t, expectedJSON, string(jsonData))
	})

	t.Run("MarshalJSON for invalid (null) NullableUUID", func(t *testing.T) {
		jsonData, err := json.Marshal(invalidNu)
		require.NoError(t, err)
		assert.Equal(t, "null", string(jsonData))
	})

	t.Run("UnmarshalJSON for valid NullableUUID", func(t *testing.T) {
		jsonInput := fmt.Sprintf("\"%s\"", validTestID.String())
		var nu types.NullableUUID
		err := json.Unmarshal([]byte(jsonInput), &nu)
		require.NoError(t, err)
		assert.True(t, nu.Valid)
		assert.Equal(t, validTestID, types.UUID(nu.UUID))
	})

	t.Run("UnmarshalJSON for invalid (null) NullableUUID", func(t *testing.T) {
		var nu types.NullableUUID
		err := json.Unmarshal([]byte("null"), &nu)
		require.NoError(t, err)
		assert.False(t, nu.Valid)
		assert.Equal(t, uuid.Nil, nu.UUID)
	})

	t.Run("UnmarshalJSON with invalid JSON UUID string", func(t *testing.T) {
		var nu types.NullableUUID
		invalidJSONString := "\"invalid-uuid-string\""
		err := json.Unmarshal([]byte(invalidJSONString), &nu)
		require.Error(t, err)
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr))
		assert.Equal(t, msg.CodeInvalid, msgErr.Code)
		assert.Contains(t, msgErr.Message, "NullableUUID must be a valid JSON UUID string or 'null'")
	})

	t.Run("UnmarshalJSON with non-string JSON", func(t *testing.T) {
		var nu types.NullableUUID
		nonStringJSON := "123"
		err := json.Unmarshal([]byte(nonStringJSON), &nu)
		require.Error(t, err)
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr))
		assert.Equal(t, msg.CodeInvalid, msgErr.Code)
		assert.Contains(t, msgErr.Message, "NullableUUID must be a valid JSON UUID string or 'null'")
	})
}

func TestNullableUUID_DatabaseEncoding(t *testing.T) {
	validTestID := mustNewTestUUID(t)
	validNu := types.NewValidNullableUUID(validTestID)
	invalidNu := types.NewNullUUID()

	t.Run("Value for valid NullableUUID", func(t *testing.T) {
		val, err := validNu.Value()
		require.NoError(t, err)
		strVal, ok := val.(string)
		require.True(t, ok)
		assert.Equal(t, validTestID.String(), strVal)
	})

	t.Run("Value for invalid (null) NullableUUID", func(t *testing.T) {
		val, err := invalidNu.Value()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("Scan valid UUID string into NullableUUID", func(t *testing.T) {
		var nu types.NullableUUID
		err := nu.Scan(validTestID.String())
		require.NoError(t, err)
		assert.True(t, nu.Valid)
		assert.Equal(t, validTestID, types.UUID(nu.UUID))
	})

	t.Run("Scan valid UUID []byte into NullableUUID", func(t *testing.T) {
		var nu types.NullableUUID
		err := nu.Scan([]byte(validTestID.String()))
		require.NoError(t, err)
		assert.True(t, nu.Valid)
		assert.Equal(t, validTestID, types.UUID(nu.UUID))
	})

	t.Run("Scan nil from DB into NullableUUID", func(t *testing.T) {
		var nu types.NullableUUID
		err := nu.Scan(nil)
		require.NoError(t, err)
		assert.False(t, nu.Valid)
	})

	t.Run("Scan incompatible type into NullableUUID", func(t *testing.T) {
		var nu types.NullableUUID
		err := nu.Scan(12345)
		require.Error(t, err)
	})
}
