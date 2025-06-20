package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewNullableTime(t *testing.T) {
	now := time.Now()
	ntValid := types.NewNullableTime(now, true)
	assert.True(t, ntValid.Valid, "NewNullableTime with valid=true should be valid")
	assert.True(t, ntValid.Time.Equal(now), "Time should be set correctly for valid NullableTime")

	ntInvalid := types.NewNullableTime(time.Time{}, false)
	assert.False(t, ntInvalid.Valid, "NewNullableTime with valid=false should be invalid")
	assert.True(t, ntInvalid.Time.IsZero(), "Time should be zero for invalid NullableTime")
}

func TestNullableTime_TimeOrZero(t *testing.T) {
	now := time.Now()
	validNt := types.NewNullableTime(now, true)
	assert.True(t, validNt.TimeOrZero().Equal(now), "TimeOrZero for valid NullableTime should return the set time")

	invalidNt := types.NewNullableTime(time.Time{}, false)
	assert.True(t, invalidNt.TimeOrZero().IsZero(), "TimeOrZero for invalid NullableTime should return zero time")
}

func TestNullableTime_Set(t *testing.T) {
	var nt types.NullableTime
	now := time.Now()
	nt.Set(now)
	assert.True(t, nt.Valid, "NullableTime should be valid after Set with non-zero time")
	assert.True(t, nt.Time.Equal(now), "Time should be correctly set")

	nt.Set(time.Time{})
	assert.False(t, nt.Valid, "NullableTime should be invalid after Set with zero time")
	assert.True(t, nt.Time.IsZero(), "Time should be zero after Set with zero time")
}

func TestNullableTime_SetNull(t *testing.T) {
	var nt types.NullableTime
	nt.Set(time.Now())
	require.True(t, nt.Valid, "Pre-condition: NullableTime should be valid")

	nt.SetNull()
	assert.False(t, nt.Valid, "NullableTime should be invalid after SetNull")
	assert.True(t, nt.Time.IsZero(), "Time should be zero after SetNull")
}

func TestNullableTime_JSONEncoding(t *testing.T) {
	specificTime, _ := time.Parse(time.RFC3339Nano, "2024-05-22T15:30:00.123Z")
	validNt := types.NewNullableTime(specificTime, true)
	invalidNt := types.NewNullableTime(time.Time{}, false)

	t.Run("Marshal valid NullableTime", func(t *testing.T) {
		jsonData, err := json.Marshal(validNt)
		require.NoError(t, err, "json.Marshal for valid NullableTime should not error")
		expectedJSON := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		assert.JSONEq(t, expectedJSON, string(jsonData), "JSON for valid NullableTime is not as expected")
	})

	t.Run("Marshal invalid (null) NullableTime", func(t *testing.T) {
		jsonData, err := json.Marshal(invalidNt)
		require.NoError(t, err, "json.Marshal for invalid NullableTime should not error")
		assert.Equal(t, "null", string(jsonData), "JSON for invalid NullableTime should be 'null'")
	})

	t.Run("Unmarshal valid NullableTime", func(t *testing.T) {
		jsonInput := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		var nt types.NullableTime
		err := json.Unmarshal([]byte(jsonInput), &nt)
		require.NoError(t, err, "json.Unmarshal for valid JSON string should not error")
		assert.True(t, nt.Valid, "NullableTime should be valid after unmarshaling valid time string")
		assert.True(t, nt.Time.Equal(specificTime), "Unmarshaled time does not match original")
	})

	t.Run("Unmarshal invalid (null) NullableTime", func(t *testing.T) {
		var nt types.NullableTime
		err := json.Unmarshal([]byte("null"), &nt)
		require.NoError(t, err, "json.Unmarshal for 'null' should not error for NullableTime")
		assert.False(t, nt.Valid, "NullableTime should be invalid after unmarshaling 'null'")
	})

	t.Run("Unmarshal invalid JSON for NullableTime", func(t *testing.T) {
		var nt types.NullableTime
		invalidJSON := `"invalid-date-string"`
		err := json.Unmarshal([]byte(invalidJSON), &nt)
		require.Error(t, err, "json.Unmarshal for invalid date string should error")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
		assert.Contains(t, msgErr.Message, "NullableTime must be a valid JSON timestamp or 'null'", "Error message content mismatch")
	})
}

func TestNullableTime_DatabaseEncoding(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	validNt := types.NewNullableTime(now, true)
	invalidNt := types.NewNullableTime(time.Time{}, false)

	t.Run("Value for valid NullableTime", func(t *testing.T) {
		val, err := validNt.Value()
		require.NoError(t, err, "Value() for valid NullableTime should not error")
		dbTime, ok := val.(time.Time)
		require.True(t, ok, "DB value should be of type time.Time")
		assert.True(t, dbTime.Equal(now), "DB value time does not match original")
	})

	t.Run("Value for invalid (null) NullableTime", func(t *testing.T) {
		val, err := invalidNt.Value()
		require.NoError(t, err, "Value() for invalid NullableTime should not error")
		assert.Nil(t, val, "DB value for invalid NullableTime should be nil")
	})

	t.Run("Scan valid time.Time into NullableTime", func(t *testing.T) {
		var nt types.NullableTime
		err := nt.Scan(now)
		require.NoError(t, err, "Scan() with valid time.Time should not error")
		assert.True(t, nt.Valid, "NullableTime should be valid after Scan")
		assert.True(t, nt.Time.Equal(now), "Scanned time does not match original")
	})

	t.Run("Scan nil from DB into NullableTime", func(t *testing.T) {
		var nt types.NullableTime
		err := nt.Scan(nil)
		require.NoError(t, err, "Scan() with nil DB value should not error")
		assert.False(t, nt.Valid, "NullableTime should be invalid after Scan with nil")
	})

	t.Run("Scan incompatible type into NullableTime", func(t *testing.T) {
		var nt types.NullableTime
		err := nt.Scan("not-a-time-value-at-all")
		assert.Error(t, err, "Scan() with incompatible type should error")
	})
}

func TestDeletedAt_Methods(t *testing.T) {
	t.Run("NewDeletedAtNow", func(t *testing.T) {
		da := types.NewDeletedAtNow()
		assert.True(t, da.Valid, "NewDeletedAtNow should result in a valid DeletedAt")
		assert.WithinDuration(t, time.Now(), da.Time, time.Second, "Time for NewDeletedAtNow should be current")
	})

	t.Run("NewNilDeletedAt", func(t *testing.T) {
		da := types.NewNilDeletedAt()
		assert.False(t, da.Valid, "NewNilDeletedAt should result in an invalid DeletedAt")
	})

	t.Run("SetNow for DeletedAt", func(t *testing.T) {
		var da types.DeletedAt
		da.SetNow()
		assert.True(t, da.Valid, "SetNow should make DeletedAt valid")
		assert.WithinDuration(t, time.Now(), da.Time, time.Second, "Time after SetNow should be current")
	})

	t.Run("JSON Encoding for DeletedAt", func(t *testing.T) {
		daValid := types.NewDeletedAtNow()
		daValid.Time = daValid.Time.UTC().Truncate(time.Microsecond)

		jsonData, err := json.Marshal(daValid)
		require.NoError(t, err, "json.Marshal for valid DeletedAt should not error")

		var newDa types.DeletedAt
		err = json.Unmarshal(jsonData, &newDa)
		require.NoError(t, err, "json.Unmarshal for valid DeletedAt JSON should not error")

		assert.True(t, newDa.Valid, "Unmarshaled DeletedAt should be valid")
		assert.True(t, daValid.Time.Equal(newDa.Time), fmt.Sprintf("Expected instant %v, got %v after JSON roundtrip", daValid.Time, newDa.Time))

		daNil := types.NewNilDeletedAt()
		jsonDataNil, err := json.Marshal(daNil)
		require.NoError(t, err, "json.Marshal for nil DeletedAt should not error")
		assert.Equal(t, "null", string(jsonDataNil), "JSON for nil DeletedAt should be 'null'")
	})
}

func TestArchivedAt_Methods(t *testing.T) {
	t.Run("NewArchivedAtNow", func(t *testing.T) {
		aa := types.NewArchivedAtNow()
		assert.True(t, aa.Valid, "NewArchivedAtNow should result in a valid ArchivedAt")
		assert.WithinDuration(t, time.Now(), aa.Time, time.Second, "Time for NewArchivedAtNow should be current")
	})

	t.Run("NewNilArchivedAt", func(t *testing.T) {
		aa := types.NewNilArchivedAt()
		assert.False(t, aa.Valid, "NewNilArchivedAt should result in an invalid ArchivedAt")
	})

	t.Run("SetNow for ArchivedAt", func(t *testing.T) {
		var aa types.ArchivedAt
		aa.SetNow()
		assert.True(t, aa.Valid, "SetNow should make ArchivedAt valid")
		assert.WithinDuration(t, time.Now(), aa.Time, time.Second, "Time after SetNow should be current")
	})

	t.Run("JSON Encoding for ArchivedAt", func(t *testing.T) {
		aaValid := types.NewArchivedAtNow()
		aaValid.Time = aaValid.Time.UTC().Truncate(time.Microsecond)

		jsonData, err := json.Marshal(aaValid)
		require.NoError(t, err, "json.Marshal for valid ArchivedAt should not error")

		var newAa types.ArchivedAt
		err = json.Unmarshal(jsonData, &newAa)
		require.NoError(t, err, "json.Unmarshal for valid ArchivedAt JSON should not error")

		assert.True(t, newAa.Valid, "Unmarshaled ArchivedAt should be valid")
		assert.True(t, aaValid.Time.Equal(newAa.Time), fmt.Sprintf("Expected instant %v, got %v after JSON roundtrip", aaValid.Time, newAa.Time))

		aaNil := types.NewNilArchivedAt()
		jsonDataNil, err := json.Marshal(aaNil)
		require.NoError(t, err, "json.Marshal for nil ArchivedAt should not error")
		assert.Equal(t, "null", string(jsonDataNil), "JSON for nil ArchivedAt should be 'null'")
	})
}
