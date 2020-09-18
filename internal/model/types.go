//go:generate mockgen -destination mock_model/modeler_mock.go humanitec.io/resources/driver-aws-external/internal/model Modeler

package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound indicates that the resource could not be found.
var ErrNotFound = errors.New("not found")

// Model is the underlying type for the entire model.
type model struct {
	*sql.DB
}

// Modeler provides an interface which can be used to mock the model
type Modeler interface {
	InsertOrUpdateResourceMetadata(m ResourceMetadata) error
	SelectResourceMetadata(id string) (ResourceMetadata, bool, error)
	DeleteResourceMetadata(id string, deletedAt time.Time) error
}

// ResourceMetadata is metadata held of a resource
type ResourceMetadata struct {
	ID        string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
	Params    map[string]interface{}
	Data      map[string]interface{}
}

func AsJSON(obj interface{}) *persisableJSON {
	return &persisableJSON{obj}
}

// A used to persist things as JSON
type persisableJSON struct {
	value interface{}
}

// Provide a way for arbitrary objects to implement the driver.Valuer interface.
func (j persisableJSON) Value() (driver.Value, error) {
	return json.Marshal(j.value)
}

// Provide a way for arbitrary objects to implement the sql.Scanner interface.
func (j *persisableJSON) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &j.value)
}
