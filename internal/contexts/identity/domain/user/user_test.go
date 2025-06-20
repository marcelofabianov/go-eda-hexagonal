package user_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/hasher"
	porthasher "github.com/marcelofabianov/redtogreen/internal/platform/port/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type mockHasher struct {
	ShouldFail bool
}

func (m *mockHasher) Hash(data string) (string, error) {
	if m.ShouldFail {
		return "", errors.New("mock hasher failed")
	}
	h, _ := hasher.NewHasher().Hash(data)
	return h, nil
}

func (m *mockHasher) Compare(data, encodedHash string) (bool, error) {
	return hasher.NewHasher().Compare(data, encodedHash)
}

func TestNewUser(t *testing.T) {
	h := &mockHasher{ShouldFail: false}

	testCases := []struct {
		name          string
		input         user.NewUserInput
		hasher        porthasher.Hasher
		expectError   bool
		errorContains string
	}{
		{
			name: "Success: create user with valid data",
			input: user.NewUserInput{
				Name:     "Marcelo Fabiano",
				Email:    "marcelo@example.com",
				Phone:    "+5562999998888",
				Password: "ValidPassword123!",
			},
			hasher:      h,
			expectError: false,
		},
		{
			name: "Failure: empty name",
			input: user.NewUserInput{
				Name:     " ",
				Email:    "marcelo@example.com",
				Phone:    "+5562999998888",
				Password: "ValidPassword123!",
			},
			hasher:        h,
			expectError:   true,
			errorContains: user.ErrUserNameRequired,
		},
		{
			name: "Failure: invalid email",
			input: user.NewUserInput{
				Name:     "Marcelo Fabiano",
				Email:    "invalid-email",
				Phone:    "+5562999998888",
				Password: "ValidPassword123!",
			},
			hasher:        h,
			expectError:   true,
			errorContains: "has an invalid format",
		},
		{
			name: "Failure: weak password",
			input: user.NewUserInput{
				Name:     "Marcelo Fabiano",
				Email:    "marcelo@example.com",
				Phone:    "+5562999998888",
				Password: "weak",
			},
			hasher:        h,
			expectError:   true,
			errorContains: "must be at least 10 characters long",
		},
		{
			name: "Failure: hasher returns an error",
			input: user.NewUserInput{
				Name:     "Marcelo Fabiano",
				Email:    "marcelo@example.com",
				Phone:    "+5562999998888",
				Password: "ValidPassword123!",
			},
			hasher:        &mockHasher{ShouldFail: true},
			expectError:   true,
			errorContains: "An unexpected internal error occurred.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newUser, err := user.NewUser(tc.input, tc.hasher)

			if tc.expectError {
				require.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain the expected text")
			} else {
				require.NoError(t, err, "Expected no error but got: %v", err)
				assert.NotNil(t, newUser, "New user should not be nil on success")
				assert.Equal(t, tc.input.Name, newUser.Name, "Name should be set correctly")
				assert.Equal(t, tc.input.Email, newUser.Email.String(), "Email should be set correctly")
				assert.False(t, newUser.Password.IsEmpty(), "Password hash should not be empty")
				assert.Equal(t, types.Version(1), newUser.Version, "Initial version should be 1")
			}
		})
	}
}

func TestUser_Update(t *testing.T) {
	h := hasher.NewHasher()
	u, err := user.NewUser(user.NewUserInput{
		Name: "Original Name", Email: "original@example.com", Phone: "+5562999998888", Password: "ValidPassword123!",
	}, h)
	require.NoError(t, err, "Setup: failed to create initial user for update tests")

	originalUpdatedAt := u.UpdatedAt
	originalVersion := u.Version

	t.Run("Success: update user name", func(t *testing.T) {
		updateInput := user.UpdateUserInput{Name: "Updated Name"}
		err := u.Update(updateInput)

		require.NoError(t, err, "Updating name should not produce an error")
		assert.Equal(t, "Updated Name", u.Name, "Name should be updated")
		assert.NotEqual(t, originalUpdatedAt, u.UpdatedAt, "UpdatedAt should be modified")
		assert.Equal(t, originalVersion+1, u.Version, "Version should be incremented by one after an update")
	})

	t.Run("Failure: update with invalid email", func(t *testing.T) {
		updateInput := user.UpdateUserInput{Email: "invalid-email"}
		err := u.Update(updateInput)
		require.Error(t, err, "Updating with an invalid email should produce an error")
	})

	t.Run("Success: no changes applied", func(t *testing.T) {
		u.Name = "Original Name"
		updateInput := user.UpdateUserInput{}
		err := u.Update(updateInput)

		require.NoError(t, err, "Update with no changes should not produce an error")
		assert.Equal(t, "Original Name", u.Name, "Name should remain unchanged")
	})
}

func TestUser_Lifecycle(t *testing.T) {
	h := hasher.NewHasher()
	u, err := user.NewUser(user.NewUserInput{
		Name: "Lifecycle User", Email: "lifecycle@example.com", Phone: "+5562999998888", Password: "ValidPassword123!",
	}, h)
	require.NoError(t, err, "Setup: failed to create user for lifecycle tests")

	t.Run("Archiving", func(t *testing.T) {
		assert.False(t, u.IsArchived(), "User should not be archived initially")

		u.Archive()
		assert.True(t, u.IsArchived(), "User should be archived after calling Archive()")
		assert.False(t, u.ArchivedAt.IsNullable(), "ArchivedAt should have a value")

		updatedAtBefore := u.UpdatedAt
		u.Archive()
		assert.Equal(t, updatedAtBefore, u.UpdatedAt, "Calling Archive on an already archived user should be idempotent")
	})

	t.Run("Unarchiving", func(t *testing.T) {
		u.Unarchive()
		assert.False(t, u.IsArchived(), "User should not be archived after calling Unarchive()")
		assert.True(t, u.ArchivedAt.IsNullable(), "ArchivedAt should be null")
	})
}

func TestUser_ComparePassword(t *testing.T) {
	h := hasher.NewHasher()
	plainPasswordStr := "ValidPassword123!@"

	u, err := user.NewUser(user.NewUserInput{
		Name: "Compare User", Email: "compare@example.com", Phone: "+5562999998888", Password: plainPasswordStr,
	}, h)
	require.NoError(t, err, "Setup: failed to create user for password comparison test")

	validPassword, _ := types.NewPassword(plainPasswordStr)
	wrongPassword, _ := types.NewPassword("WrongPassword123!")

	t.Run("Success: correct password", func(t *testing.T) {
		match, err := u.ComparePassword(validPassword, h)
		require.NoError(t, err, "Comparison should not return an error")
		assert.True(t, match, "Comparison with correct password should return true")
	})

	t.Run("Failure: incorrect password", func(t *testing.T) {
		match, err := u.ComparePassword(wrongPassword, h)
		require.NoError(t, err, "Comparison should not return an error")
		assert.False(t, match, "Comparison with incorrect password should return false")
	})
}
