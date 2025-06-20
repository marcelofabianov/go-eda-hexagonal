package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/marcelofabianov/redtogreen/internal/platform/port/hasher"
)

type HashedPassword string

func NewHashedPassword(hash string) HashedPassword {
	return HashedPassword(hash)
}

func (hp HashedPassword) String() string {
	return string(hp)
}

func (hp HashedPassword) IsEmpty() bool {
	return hp.String() == ""
}

func (hp HashedPassword) Compare(plaintextPassword Password, h hasher.Hasher) (bool, error) {
	if hp.IsEmpty() || plaintextPassword.IsEmpty() {
		return false, nil
	}
	return h.Compare(plaintextPassword.String(), hp.String())
}

func (hp HashedPassword) MarshalJSON() ([]byte, error) {
	return json.Marshal(hp.String())
}

func (hp *HashedPassword) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*hp = NewHashedPassword(s)
	return nil
}

func (hp HashedPassword) Value() (driver.Value, error) {
	if hp.IsEmpty() {
		return nil, nil
	}
	return hp.String(), nil
}

func (hp *HashedPassword) Scan(src interface{}) error {
	if src == nil {
		*hp = ""
		return nil
	}

	var hash string
	switch sval := src.(type) {
	case string:
		hash = sval
	case []byte:
		hash = string(sval)
	default:
		return fmt.Errorf("incompatible type (%T) for HashedPassword scan", src)
	}

	*hp = NewHashedPassword(hash)
	return nil
}
