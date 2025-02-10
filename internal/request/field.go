package request

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhaifley/apigo/internal/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/yaml.v3"
)

// FieldString values represent strings tolerant of JSON inputs.
type FieldString struct {
	Set   bool
	Valid bool
	Value string
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldString) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = ""

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case string:
		f.Value = tv
	case float64:
		f.Value = strconv.FormatFloat(tv, 'f', -1, 64)
	case int64:
		f.Value = strconv.FormatInt(tv, 10)
	case bool:
		f.Value = strconv.FormatBool(tv)
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into string",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldString) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldString) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldString) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldString) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = ""

	switch v := src.(type) {
	case []byte:
		f.Value = string(v)
	case string:
		f.Value = v
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into string", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldString) String() string {
	return f.Value
}

// FieldInt64 values represent integers tolerant of JSON inputs.
type FieldInt64 struct {
	Set   bool
	Valid bool
	Value int64
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldInt64) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case string:
		i, err := strconv.ParseInt(tv, 10, 64)
		if err != nil {
			n, nErr := strconv.ParseFloat(tv, 64)
			if nErr != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse JSON string into int64",
					"json", string(b),
					"string", tv)
			}

			i = int64(n)
		}

		f.Value = i
	case float64:
		f.Value = int64(tv)
	case int64:
		f.Value = int64(tv)
	case bool:
		if tv {
			f.Value = 1
		} else {
			f.Value = 0
		}
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into int64",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldInt64) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldInt64) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if value == nil {
		f.Valid = false

		return nil
	}

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldInt64) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldInt64) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	switch v := src.(type) {
	case int64:
		f.Value = v
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into int64", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldInt64) String() string {
	return strconv.FormatInt(f.Value, 10)
}

// FieldFloat64 values represent floats tolerant of JSON inputs.
type FieldFloat64 struct {
	Set   bool
	Valid bool
	Value float64
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldFloat64) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = 0.0

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case string:
		n, err := strconv.ParseFloat(tv, 64)
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse JSON string into float64",
				"json", string(b),
				"string", tv)
		}

		f.Value = n
	case float64:
		f.Value = tv
	case int64:
		f.Value = float64(tv)
	case bool:
		if tv {
			f.Value = 1.0
		} else {
			f.Value = 0.0
		}
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into float64",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldFloat64) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldFloat64) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldFloat64) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldFloat64) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	switch v := src.(type) {
	case float64:
		f.Value = v
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into float64", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldFloat64) String() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

// FieldBool values represent booleans tolerant of JSON inputs.
type FieldBool struct {
	Set   bool
	Valid bool
	Value bool
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldBool) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = false

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case string:
		bv, err := strconv.ParseBool(tv)
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse JSON string into bool",
				"json", string(b),
				"string", tv)
		}

		f.Value = bv
	case float64:
		if tv == 0 {
			f.Value = false
		} else {
			f.Value = true
		}
	case int64:
		if tv == 0 {
			f.Value = false
		} else {
			f.Value = true
		}
	case bool:
		f.Value = tv
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into bool",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldBool) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldBool) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldBool) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldBool) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = false

	switch v := src.(type) {
	case bool:
		f.Value = v
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into bool", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldBool) String() string {
	return strconv.FormatBool(f.Value)
}

// FieldTime values represent timestamps tolerant of JSON inputs.
type FieldTime struct {
	Set   bool
	Valid bool
	Value int64
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldTime) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case string:
		i, err := strconv.ParseInt(tv, 10, 64)
		if err != nil {
			t, tErr := time.Parse(time.RFC3339, tv)
			if tErr != nil {
				return errors.Wrap(tErr, errors.ErrInvalidRequest,
					"unable to parse JSON string into timestamp",
					"json", string(b),
					"string", tv)
			}

			i = t.Unix()
		}

		f.Value = i
	case float64:
		f.Value = int64(tv)
	case int64:
		f.Value = int64(tv)
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into timestamp",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldTime) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldTime) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldTime) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldTime) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	switch v := src.(type) {
	case time.Time:
		f.Value = v.Unix()
	case int64:
		f.Value = v
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into int64", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldTime) String() string {
	return strconv.FormatInt(f.Value, 10)
}

// cardinality returns the number of elements in an array of dimensions size.
func cardinality(dimensions []pgtype.ArrayDimension) int {
	if len(dimensions) == 0 {
		return 0
	}

	elementCount := int(dimensions[0].Length)

	for _, d := range dimensions[1:] {
		elementCount *= int(d.Length)
	}

	return elementCount
}

// FieldStringArray values represent string arrays tolerant of JSON inputs.
type FieldStringArray struct {
	Set   bool
	Valid bool
	Value []string
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldStringArray) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = nil

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case []any:
		for _, vv := range tv {
			if sv, ok := vv.(string); ok {
				f.Value = append(f.Value, sv)
			} else {
				return errors.New(errors.ErrInvalidRequest,
					"unable to parse JSON array into []string",
					"json", string(b))
			}
		}
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into []string",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldStringArray) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldStringArray) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldStringArray) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Dimensions supports the pgtype ArrayGetter interface.
func (f FieldStringArray) Dimensions() []pgtype.ArrayDimension {
	if !f.Set || !f.Valid {
		return nil
	}

	return []pgtype.ArrayDimension{{Length: int32(len(f.Value)), LowerBound: 1}}
}

// Index supports the pgtype ArrayGetter interface.
func (f FieldStringArray) Index(i int) any {
	if !f.Set || !f.Valid {
		return nil
	}

	return f.Value[i]
}

// IndexType supports the pgtype ArrayGetter interface.
func (f FieldStringArray) IndexType() any {
	var el string

	return el
}

// SetDimensions supports the pgtype ArraySetter interface.
func (f *FieldStringArray) SetDimensions(dimensions []pgtype.ArrayDimension,
) error {
	f.Set = true
	f.Valid = true

	if dimensions == nil {
		f.Valid = false
		f.Value = nil

		return nil
	}

	elementCount := cardinality(dimensions)

	f.Value = make([]string, elementCount)

	return nil
}

// ScanIndex supports the pgtype ArraySetter interface.
func (f FieldStringArray) ScanIndex(i int) any {
	if !f.Set || !f.Valid {
		return nil
	}

	return &(f.Value[i])
}

// ScanIndexType supports the pgtype ArraySetter interface.
func (f FieldStringArray) ScanIndexType() any {
	return new(string)
}

// String returns the value as a string.
func (f *FieldStringArray) String() string {
	return strings.Join(f.Value, " ")
}

// FieldInt64Array values represent string arrays tolerant of JSON inputs.
type FieldInt64Array struct {
	Set   bool
	Valid bool
	Value []int64
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldInt64Array) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = nil

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch tv := v.(type) {
	case []any:
		for _, sv := range tv {
			switch vv := sv.(type) {
			case int64:
				f.Value = append(f.Value, vv)
			case float64:
				f.Value = append(f.Value, int64(vv))
			default:
				return errors.New(errors.ErrInvalidRequest,
					"unable to parse JSON array into []string",
					"json", string(b))
			}
		}
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse JSON into []string",
			"json", string(b))
	}

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldInt64Array) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldInt64Array) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldInt64Array) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Dimensions supports the pgtype ArrayGetter interface.
func (f FieldInt64Array) Dimensions() []pgtype.ArrayDimension {
	if !f.Set || !f.Valid {
		return nil
	}

	return []pgtype.ArrayDimension{{Length: int32(len(f.Value)), LowerBound: 1}}
}

// Index supports the pgtype ArrayGetter interface.
func (f FieldInt64Array) Index(i int) any {
	if !f.Set || !f.Valid {
		return nil
	}

	return f.Value[i]
}

// IndexType supports the pgtype ArrayGetter interface.
func (f FieldInt64Array) IndexType() any {
	var el int64

	return el
}

// SetDimensions supports the pgtype ArraySetter interface.
func (f *FieldInt64Array) SetDimensions(dimensions []pgtype.ArrayDimension,
) error {
	f.Set = true
	f.Valid = true

	if dimensions == nil {
		f.Valid = false
		f.Value = nil

		return nil
	}

	elementCount := cardinality(dimensions)

	f.Value = make([]int64, elementCount)

	return nil
}

// ScanIndex supports the pgtype ArraySetter interface.
func (f FieldInt64Array) ScanIndex(i int) any {
	if !f.Set || !f.Valid {
		return nil
	}

	return &(f.Value[i])
}

// ScanIndexType supports the pgtype ArraySetter interface.
func (f FieldInt64Array) ScanIndexType() any {
	return new(int64)
}

// String returns the value as a string.
func (f *FieldInt64Array) String() string {
	return fmt.Sprintf("%v", f.Value)
}

// FieldJSON values represent unparsed JSON objects.
type FieldJSON struct {
	Set   bool
	Valid bool
	Value map[string]any
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldJSON) UnmarshalJSON(b []byte) error {
	f.Set = true

	if err := json.Unmarshal(b, &f.Value); err != nil {
		return err
	}

	f.Valid = (f.Value != nil)

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldJSON) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value)
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldJSON) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldJSON) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldJSON) Scan(src any) error {
	f.Set = true

	switch v := src.(type) {
	case []byte:
		if err := f.UnmarshalJSON(v); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to scan value into JSON object",
				"value", string(v))
		}
	case string:
		if err := f.UnmarshalJSON([]byte(v)); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to scan value into JSON object",
				"value", v)
		}
	case nil:
		f.Value = nil
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into JSON object", v))
	}

	f.Valid = (f.Value != nil)

	return nil
}

// String returns the value as a string.
func (f *FieldJSON) String() string {
	if b, err := json.Marshal(f.Value); err == nil {
		return string(b)
	}

	return "{}"
}

// FieldDuration values represent integers tolerant of JSON inputs.
type FieldDuration struct {
	Set   bool
	Valid bool
	Value time.Duration
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (f *FieldDuration) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	var v any

	if err := json.Unmarshal(b, &v); err != nil {
		return errors.Wrap(err, errors.ErrInvalidRequest,
			"unable to parse JSON into duration",
			"json", string(b))
	}

	switch val := v.(type) {
	case float64:
		if val > 10000000000 {
			f.Value = time.Duration(val)
		} else {
			f.Value = time.Duration(time.Second * time.Duration(val))
		}

		return nil
	case string:
		var err error

		f.Value, err = time.ParseDuration(val)
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse duration",
				"value", v)
		}

		return nil
	case nil:
		f.Valid = false

		return nil
	default:
		return errors.New(errors.ErrInvalidRequest,
			"invalid duration",
			"value", v)
	}
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (f *FieldDuration) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(f.Value.String())
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (f *FieldDuration) UnmarshalYAML(value *yaml.Node) error {
	f.Set = true
	f.Valid = true

	if err := value.Decode(&f.Value); err != nil {
		return err
	}

	return nil
}

// MarshalYAML encodes a this value into a YAML format byte slice.
func (f FieldDuration) MarshalYAML() (any, error) {
	if !f.Set || !f.Valid {
		return nil, nil
	}

	return f.Value, nil
}

// Scan allows this value to be used in database/sql scan functions.
func (f *FieldDuration) Scan(src any) error {
	f.Set = true
	f.Valid = true
	f.Value = 0

	switch v := src.(type) {
	case int64:
		f.Value = time.Duration(v)
	case nil:
		f.Valid = false
	default:
		return errors.New(errors.ErrDatabase,
			fmt.Sprintf("unable to scan value of type %T into duration", v))
	}

	return nil
}

// String returns the value as a string.
func (f *FieldDuration) String() string {
	return f.Value.String()
}

// SetField adds the name and value for a field to the provided set lists.
func SetField(name string, field any,
	sets *[]string, params *[]any,
) {
	if sets == nil || params == nil {
		return
	}

	switch f := field.(type) {
	case FieldString:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldInt64:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldFloat64:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldBool:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldTime:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldStringArray:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldInt64Array:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldJSON:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				b, err := json.Marshal(f.Value)
				if err == nil {
					*params = append(*params, b)
				} else {
					*params = append(*params, []byte("{}"))
				}
			} else {
				*params = append(*params, nil)
			}
		}
	case FieldDuration:
		if f.Set {
			*sets = append(*sets, name)

			if f.Valid {
				*params = append(*params, f.Value)
			} else {
				*params = append(*params, nil)
			}
		}
	}
}
