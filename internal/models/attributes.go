package models

import (
	"acrocuit/schema"
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrAttributesNotObject = errors.New("attributes must be a JSON object")
	ErrAttributesTooLarge  = fmt.Errorf("attributes must not exceed %d bytes", schema.AttributesMaxBytes)
)

var emptyAttributes = Attributes("{}")

type Attributes []byte

func (a Attributes) OrEmpty() Attributes {
	if a == nil {
		return emptyAttributes
	}
	return a
}

func (a Attributes) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	return a, nil
}

func (a *Attributes) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = nil
		return nil
	}

	var compacted bytes.Buffer
	if err := json.Compact(&compacted, data); err != nil {
		return err
	}
	if compacted.Bytes()[0] != '{' {
		return ErrAttributesNotObject
	}
	if compacted.Len() > schema.AttributesMaxBytes {
		return ErrAttributesTooLarge
	}

	*a = compacted.Bytes()
	return nil
}

func (a *Attributes) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = nil
	case string:
		*a = Attributes(v)
	case []byte:
		*a = bytes.Clone(v)
	default:
		return fmt.Errorf("cannot scan %T into Attributes", src)
	}
	return nil
}

func (a Attributes) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return string(a), nil
}
