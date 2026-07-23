package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type MapStrInt64 map[string]int64

func (m *MapStrInt64) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func (m *MapStrInt64) Scan(value interface{}) error {
	if value == nil {
		*m = MapStrInt64{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("MapStrInt64: Scan source is not []byte")
	}
	return json.Unmarshal(bytes, m)
}
