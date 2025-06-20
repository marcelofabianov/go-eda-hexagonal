package user

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/port/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

const (
	UserNameMaxLength = 100
)

const (
	ErrUserNameRequired                  = "Name cannot be empty."
	ErrUserNameTooLong                   = "Name is too long (max 100 characters)."
	ErrUserEmailRequired                 = "Email cannot be empty."
	ErrUserPhoneRequired                 = "Phone cannot be empty."
	ErrUserPreferencesInvalidJSON        = "Preferences field contains invalid JSON."
	ErrUserEmailInvalid                  = "Invalid email provided."
	ErrUserPhoneInvalid                  = "Invalid phone number provided."
	ErrUserPasswordRequired              = "Password is required."
	ErrUserPreferencesNotValidJSON       = "User preferences is not valid JSON."
	ErrUserEmailInvalidUpdate            = "Invalid email provided for update."
	ErrUserPhoneInvalidUpdate            = "Invalid phone number provided for update."
	ErrUserPreferencesNotValidJSONUpdate = "User preferences is not valid JSON during update."
	ErrUserOperationGenerateUUID         = "Failed to generate UUID for user."
	ErrUserOperationHashPassword         = "Failed to hash user password."
	ErrUserAlreadyExists                 = "A user with the given email or phone already exists."
)

type NewUserInput struct {
	Name        string
	Email       string
	Phone       string
	Password    string
	Preferences json.RawMessage `json:"preferences,omitempty"`
}

type UpdateUserInput struct {
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Preferences json.RawMessage `json:"preferences,omitempty"`
}

type FromUserInput struct {
	ID          types.UUID
	Name        string
	Email       types.Email
	Phone       types.Phone
	Password    types.HashedPassword
	Preferences json.RawMessage
	CreatedAt   types.CreatedAt
	UpdatedAt   types.UpdatedAt
	Version     types.Version
	ArchivedAt  types.ArchivedAt
	DeletedAt   types.DeletedAt
}

type User struct {
	ID          types.UUID           `json:"id" db:"id"`
	Name        string               `json:"name" db:"name"`
	Email       types.Email          `json:"email" db:"email"`
	Phone       types.Phone          `json:"phone" db:"phone"`
	Password    types.HashedPassword `json:"-" db:"password"`
	Preferences json.RawMessage      `json:"preferences,omitempty" db:"preferences"`
	CreatedAt   types.CreatedAt      `json:"created_at" db:"created_at"`
	UpdatedAt   types.UpdatedAt      `json:"updated_at" db:"updated_at"`
	Version     types.Version        `json:"version" db:"version"`
	ArchivedAt  types.ArchivedAt     `json:"archived_at,omitempty" db:"archived_at"`
	DeletedAt   types.DeletedAt      `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (u *User) ComparePassword(plaintextPassword types.Password, h hasher.Hasher) (bool, error) {
	return u.Password.Compare(plaintextPassword, h)
}

func (u *User) validate() error {
	nameTrimmed := strings.TrimSpace(u.Name)
	if nameTrimmed == "" {
		return msg.NewValidationError(nil, map[string]any{"name": u.Name}, ErrUserNameRequired)
	}
	if utf8.RuneCountInString(nameTrimmed) > UserNameMaxLength {
		return msg.NewValidationError(nil, map[string]any{"name": u.Name, "max_length": UserNameMaxLength}, ErrUserNameTooLong)
	}
	u.Name = nameTrimmed

	if u.Email.IsEmpty() {
		return msg.NewValidationError(nil, map[string]any{"email": u.Email.String()}, ErrUserEmailRequired)
	}
	if u.Phone.IsEmpty() {
		return msg.NewValidationError(nil, map[string]any{"phone": u.Phone.String()}, ErrUserPhoneRequired)
	}

	if u.Preferences != nil && !json.Valid(u.Preferences) {
		return msg.NewValidationError(nil, map[string]any{"field": "Preferences"}, ErrUserPreferencesInvalidJSON)
	}

	return nil
}

func NewUser(input NewUserInput, h hasher.Hasher) (*User, error) {
	if input.Password == "" {
		return nil, msg.NewValidationError(nil, map[string]any{"field": "Password"}, ErrUserPasswordRequired)
	}

	id, err := types.NewUUID()
	if err != nil {
		return nil, msg.NewInternalError(err, map[string]any{"operation": ErrUserOperationGenerateUUID})
	}

	var msgErr *msg.MessageError

	email, err := types.NewEmail(input.Email)
	if err != nil {
		if errors.As(err, &msgErr) {
			return nil, msgErr.WithContext("field", "Email")
		}
		return nil, msg.NewValidationError(err, map[string]any{"field": "Email", "input_email": input.Email}, ErrUserEmailInvalid)
	}

	phone, err := types.NewPhone(input.Phone)
	if err != nil {
		if errors.As(err, &msgErr) {
			return nil, msgErr.WithContext("field", "Phone")
		}
		return nil, msg.NewValidationError(err, map[string]any{"field": "Phone", "input_phone": input.Phone}, ErrUserPhoneInvalid)
	}

	password, err := types.NewPassword(input.Password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := h.Hash(password.String())
	if err != nil {
		return nil, msg.NewInternalError(err, map[string]any{"operation": ErrUserOperationHashPassword})
	}

	var preferencesValue json.RawMessage
	if input.Preferences != nil {
		if !json.Valid(input.Preferences) {
			return nil, msg.NewValidationError(nil, map[string]any{"field": "Preferences"}, ErrUserPreferencesNotValidJSON)
		}
		if string(input.Preferences) != "null" && len(input.Preferences) > 0 {
			preferencesValue = input.Preferences
		}
	}

	currentTime := time.Now()

	user := &User{
		ID:          id,
		Name:        input.Name,
		Email:       email,
		Phone:       phone,
		Password:    types.NewHashedPassword(hashedPassword),
		Preferences: preferencesValue,
		CreatedAt:   types.CreatedAt(currentTime),
		UpdatedAt:   types.UpdatedAt(currentTime),
		Version:     types.NewVersion(),
		ArchivedAt:  types.NewNilArchivedAt(),
		DeletedAt:   types.NewNilDeletedAt(),
	}

	if err := user.validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func FromUser(input FromUserInput) *User {
	return &User{
		ID:          input.ID,
		Name:        input.Name,
		Email:       input.Email,
		Phone:       input.Phone,
		Password:    input.Password,
		Preferences: input.Preferences,
		CreatedAt:   input.CreatedAt,
		UpdatedAt:   input.UpdatedAt,
		Version:     input.Version,
		ArchivedAt:  input.ArchivedAt,
		DeletedAt:   input.DeletedAt,
	}
}

func (u *User) Update(input UpdateUserInput) error {
	changed := false
	var msgErr *msg.MessageError

	if input.Name != "" {
		u.Name = input.Name
		changed = true
	}

	if input.Email != "" {
		email, err := types.NewEmail(input.Email)
		if err != nil {
			if errors.As(err, &msgErr) {
				return msgErr.WithContext("field", "Email")
			}
			return msg.NewValidationError(err, map[string]any{"field": "Email", "input_email": input.Email}, ErrUserEmailInvalidUpdate)
		}
		u.Email = email
		changed = true
	}

	if input.Phone != "" {
		phone, err := types.NewPhone(input.Phone)
		if err != nil {
			if errors.As(err, &msgErr) {
				return msgErr.WithContext("field", "Phone")
			}
			return msg.NewValidationError(err, map[string]any{"field": "Phone", "input_phone": input.Phone}, ErrUserPhoneInvalidUpdate)
		}
		u.Phone = phone
		changed = true
	}

	if input.Preferences != nil {
		if !json.Valid(input.Preferences) {
			return msg.NewValidationError(nil, map[string]any{"field": "Preferences"}, ErrUserPreferencesNotValidJSONUpdate)
		}
		if string(input.Preferences) == "null" || len(input.Preferences) == 0 {
			u.Preferences = nil
		} else {
			u.Preferences = input.Preferences
		}
		changed = true
	}

	if !changed {
		return nil
	}

	if err := u.validate(); err != nil {
		return err
	}

	u.UpdatedAt = types.NewUpdatedAt()
	u.Version.Increment()
	return nil
}

func (u *User) Archive() {
	if !u.IsArchived() {
		u.ArchivedAt = types.NewArchivedAtNow()
		u.UpdatedAt = types.NewUpdatedAt()
		u.Version.Increment()
	}
}

func (u *User) Unarchive() {
	if u.IsArchived() {
		u.ArchivedAt.SetNull()
		u.UpdatedAt = types.NewUpdatedAt()
		u.Version.Increment()
	}
}

func (u *User) Delete() {
	if !u.IsDeleted() {
		u.DeletedAt = types.NewDeletedAtNow()
		u.UpdatedAt = types.NewUpdatedAt()
		u.Version.Increment()
	}
}

func (u *User) Restore() {
	if u.IsDeleted() {
		u.DeletedAt.SetNull()
		u.UpdatedAt = types.NewUpdatedAt()
		u.Version.Increment()
	}
}

func (u *User) IsArchived() bool {
	return !u.ArchivedAt.IsNullable()
}

func (u *User) IsDeleted() bool {
	return !u.DeletedAt.IsNullable()
}
