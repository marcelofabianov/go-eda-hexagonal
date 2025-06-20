package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestHashedPassword_Compare(t *testing.T) {
	hasher := hasher.NewHasher()
	validPassword, err := types.NewPassword("ValidPass123!@")
	require.NoError(t, err, "Setup: creating a valid password should not fail")

	hashed, err := hasher.Hash(validPassword.String())
	require.NoError(t, err, "Setup: hashing a valid password should not fail")
	hp := types.NewHashedPassword(hashed)

	t.Run("Success: should return true for correct password", func(t *testing.T) {
		match, err := hp.Compare(validPassword, hasher)
		require.NoError(t, err, "Compare with correct password should not produce an error")
		assert.True(t, match, "Compare should return true for the correct password")
	})

	t.Run("Failure: should return false for incorrect password", func(t *testing.T) {
		wrongPassword, err := types.NewPassword("WrongPass123!@")
		require.NoError(t, err, "Setup: creating a wrong password should not fail")

		match, err := hp.Compare(wrongPassword, hasher)
		require.NoError(t, err, "Compare with incorrect password should not produce an error")
		assert.False(t, match, "Compare should return false for an incorrect password")
	})

	t.Run("Edge Case: should return false for empty plaintext password", func(t *testing.T) {
		var emptyPassword types.Password
		match, err := hp.Compare(emptyPassword, hasher)
		require.NoError(t, err, "Compare with empty password should not produce an error")
		assert.False(t, match, "Compare with an empty password should return false")
	})
}

func TestHashedPassword_JSON(t *testing.T) {
	hashString := "argon2$some_salt$some_hash"
	hp := types.NewHashedPassword(hashString)
	expectedJSON := `"` + hashString + `"`

	t.Run("Success: MarshalJSON", func(t *testing.T) {
		data, err := json.Marshal(hp)
		require.NoError(t, err, "Marshalling HashedPassword to JSON should not fail")
		assert.JSONEq(t, expectedJSON, string(data), "Marshalled JSON should be the correct string representation")
	})

	t.Run("Success: UnmarshalJSON", func(t *testing.T) {
		var newHp types.HashedPassword
		err := json.Unmarshal([]byte(expectedJSON), &newHp)
		require.NoError(t, err, "Unmarshalling JSON to HashedPassword should not fail")
		assert.Equal(t, hp, newHp, "Unmarshalled HashedPassword should have the correct value")
	})
}

func TestHashedPassword_Database(t *testing.T) {
	t.Run("Value: should return correct driver.Value for non-empty hash", func(t *testing.T) {
		hashString := "db_hash_string"
		hp := types.NewHashedPassword(hashString)

		val, err := hp.Value()
		require.NoError(t, err, "Calling Value() on a valid HashedPassword should not produce an error")

		strVal, ok := val.(string)
		require.True(t, ok, "The driver.Value should be of type string")
		assert.Equal(t, hashString, strVal, "The driver.Value should be the underlying string")
	})

	t.Run("Value: should return nil for empty HashedPassword", func(t *testing.T) {
		hp := types.NewHashedPassword("")
		val, err := hp.Value()
		require.NoError(t, err, "Calling Value() on an empty HashedPassword should not produce an error")
		assert.Nil(t, val, "The driver.Value for an empty HashedPassword should be nil")
	})

	t.Run("Scan: should correctly scan a string", func(t *testing.T) {
		var hp types.HashedPassword
		dbValue := "scanned_hash_from_db"
		err := hp.Scan(dbValue)
		require.NoError(t, err, "Scanning a string value should not produce an error")
		assert.Equal(t, dbValue, hp.String(), "HashedPassword should contain the scanned string value")
	})

	t.Run("Scan: should correctly scan a byte slice", func(t *testing.T) {
		var hp types.HashedPassword
		dbValue := []byte("scanned_hash_from_db_bytes")
		err := hp.Scan(dbValue)
		require.NoError(t, err, "Scanning a byte slice value should not produce an error")
		assert.Equal(t, string(dbValue), hp.String(), "HashedPassword should contain the string representation of the scanned bytes")
	})

	t.Run("Scan: should correctly scan a nil value", func(t *testing.T) {
		var hp types.HashedPassword
		err := hp.Scan(nil)
		require.NoError(t, err, "Scanning a nil value should not produce an error")
		assert.True(t, hp.IsEmpty(), "HashedPassword should be empty after scanning nil")
	})

	t.Run("Scan: should return an error for incompatible type", func(t *testing.T) {
		var hp types.HashedPassword
		err := hp.Scan(12345)
		require.Error(t, err, "Scanning an incompatible type like int should produce an error")
	})
}
