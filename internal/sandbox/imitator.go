package sandbox

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/volatiletech/null/v8"
	"golang.org/x/exp/constraints"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unsafe"
)

func Merge(entity interface{}, values map[string]interface{}) error {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	for name, value := range values {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			columnName := field.Tag.Get("boil")
			columnName = strings.TrimSpace(columnName)
			if columnName == "" || columnName == "-" {
				columnName = field.Name
			}

			if columnName == name {
				fieldValue := v.Field(i)
				if !fieldValue.CanSet() {
					return fmt.Errorf("field %s cannot be set", field.Name)
				}
				newColumnValue := reflect.ValueOf(value)
				if !newColumnValue.Type().AssignableTo(fieldValue.Type()) {
					return fmt.Errorf("type mismatch: cannot assign %T to %s", value, fieldValue.Type())
				}
				fieldValue.Set(newColumnValue)
			}
		}
	}

	return nil
}

func ImitatorSql[ID constraints.Ordered, T any](
	entities map[ID]*T,
	where []*expression.Where,
	groupBy []*expression.GroupBy,
	orderBy []*expression.OrderBy,
) ([]*T, error) {
	items := make([]*ImitatorModel, 0, len(entities))
	ids := make(map[unsafe.Pointer]ID, len(entities))

	for id, entity := range entities {
		m, err := RecognizeImitatorModel(entity)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
		p := unsafe.Pointer(m)
		ids[p] = id
	}

	if len(where) > 0 {
		var err error
		items, err = ImitatorSqlWhere(items, where...)
		if err != nil {
			return nil, err
		}
	}

	if len(groupBy) > 0 {
		var err error
		items, err = ImitatorSqlGroupBy(items, groupBy...)
		if err != nil {
			return nil, err
		}
	}

	if len(orderBy) > 0 {
		err := ImitatorSqlOrderBy(items, orderBy...)
		if err != nil {
			return nil, err
		}
	}

	res := make([]*T, len(items))
	for i, item := range items {
		id, ok := ids[unsafe.Pointer(item)]
		if !ok {
			return nil, fmt.Errorf("could not find id for %v", item)
		}
		entity, ok := entities[id]
		if !ok || entity == nil {
			return nil, fmt.Errorf("could not recognize entity by %v", id)
		}
		res[i] = entity
	}

	return res, nil
}

func ImitatorSqlWhere(entities []*ImitatorModel, expressions ...*expression.Where) ([]*ImitatorModel, error) {
	result := make([]*ImitatorModel, 0, len(entities))

	for _, entity := range entities {
		needed := true
		for _, eWhere := range expressions {
			res, err := entity.Compare(eWhere.Operator, eWhere.Table, eWhere.Column, eWhere.Value)
			if err != nil {
				return nil, err
			}
			needed = res
			if !needed {
				break
			}
		}
		if needed {
			result = append(result, entity)
		}
	}

	return result, nil
}

func ImitatorSqlOrderBy(entities []*ImitatorModel, expressions ...*expression.OrderBy) (err error) {
	sort.SliceStable(entities, func(a, b int) bool {
		for _, eOrderBy := range expressions {
			valA, existsI := entities[a].GetValue(eOrderBy.Table, eOrderBy.Column)
			if !existsI {
				err = errors.Join(err, fmt.Errorf("could not find %s.%s in %v", eOrderBy.Table, eOrderBy.Column, entities[a]))
				return false
			}
			valB, existsJ := entities[b].GetValue(eOrderBy.Table, eOrderBy.Column)
			if !existsJ {
				err = errors.Join(err, fmt.Errorf("could not find %s.%s in %v", eOrderBy.Table, eOrderBy.Column, entities[b]))
				return false
			}
			if err != nil {
				return false
			}
			if valB == nil && valA != nil {
				return true
			}
			if valA == valB || valA == nil || valB == nil {
				continue
			}
			switch eOrderBy.Direction {
			case expression.Ascending:
				switch v := valA.(type) {
				case int:
					return v < valB.(int)
				case uint:
					return v < valB.(uint)
				case string:
					return strings.Compare(v, valB.(string)) < 0
				case float64:
					return v < valB.(float64)
				case time.Time:
					return v.Before(valB.(time.Time))
				case []byte:
					return bytes.Compare(v, valB.([]byte)) < 0
				case bool:
					return v == true && valB.(bool) == false
				}
			case expression.Descending:
				switch v := valA.(type) {
				case int:
					return v > valB.(int)
				case uint:
					return v < valB.(uint)
				case string:
					return strings.Compare(v, valB.(string)) > 0
				case float64:
					return v > valB.(float64)
				case time.Time:
					return v.After(valB.(time.Time))
				case []byte:
					return bytes.Compare(v, valB.([]byte)) > 0
				case bool:
					return v == false && valB.(bool) == true
				}
			}
		}
		return false
	})
	return
}

func ImitatorSqlGroupBy(entities []*ImitatorModel, expressions ...*expression.GroupBy) ([]*ImitatorModel, error) {
	tmp := make(map[interface{}]*ImitatorModel, len(entities))
	for _, entity := range entities {
		for _, eGroupBy := range expressions {
			res, exists := entity.GetValue(eGroupBy.Table, eGroupBy.Column)
			if exists {
				tmp[res] = entity
			}
		}
	}

	entities = make([]*ImitatorModel, 0, len(tmp))
	for _, entity := range tmp {
		entities = append(entities, entity)
	}

	return entities, nil
}

func RecognizeImitatorModel(entity interface{}) (*ImitatorModel, error) {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type %s of object %T", t.Kind().String(), entity)
	}

	fLength := t.NumField()

	var values ImitatorModel = make(map[string]interface{}, fLength)

	for i := 0; i < fLength; i++ {
		fType := t.Field(i)
		fValue := v.Field(i)

		columnName := fType.Tag.Get("boil")
		columnName = strings.TrimSpace(columnName)
		if columnName == "" || columnName == "-" {
			columnName = fType.Name
		}

		switch fValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values[columnName] = int(fValue.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			values[columnName] = uint(fValue.Uint())
		case reflect.Float32, reflect.Float64:
			values[columnName] = fValue.Float()
		case reflect.Bool:
			values[columnName] = fValue.Bool()
		case reflect.String:
			values[columnName] = fValue.String()
		case reflect.Slice, reflect.Array:
			switch f := fValue.Type().Elem().Kind(); f {
			case reflect.Bool:
				l := fValue.Len()
				items := make([]bool, l)
				for j := 0; j < l; j++ {
					items[j] = fValue.Index(j).Bool()
				}
				values[columnName] = items
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				l := fValue.Len()
				items := make([]int, l)
				for j := 0; j < l; j++ {
					items[j] = int(fValue.Index(j).Int())
				}
				values[columnName] = items
			case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				l := fValue.Len()
				items := make([]uint, l)
				for j := 0; j < l; j++ {
					items[j] = uint(fValue.Index(j).Uint())
				}
				values[columnName] = items
			case reflect.Float32, reflect.Float64:
				l := fValue.Len()
				items := make([]float64, l)
				for j := 0; j < l; j++ {
					items[j] = fValue.Index(j).Float()
				}
				values[columnName] = items
			case reflect.String:
				l := fValue.Len()
				items := make([]string, l)
				for j := 0; j < l; j++ {
					items[j] = fValue.Index(j).String()
				}
				values[columnName] = items
			case reflect.Uint8:
				l := fValue.Len()
				items := make([]byte, l)
				for j := 0; j < l; j++ {
					items[j] = byte(fValue.Index(j).Uint())
				}
				values[columnName] = items
			default:
				return nil, fmt.Errorf("unsupported type %s of field %s", f, columnName)
			}
		case reflect.Pointer:
			if fValue.IsNil() {
				values[columnName] = nil
			} else {
				internalDataset, err := RecognizeImitatorModel(fValue.Elem().Interface())
				if err != nil {
					return nil, fmt.Errorf("could not parse %s field: %w", columnName, err)
				}
				if columnName == "R" || columnName == "L" {
					var relation ImitatorModel = make(map[string]interface{})
					for key, val := range *internalDataset {
						key = toLowerTableName(key)
						relation[key] = val
					}
					internalDataset = &relation
				}
				values[columnName] = internalDataset
			}
		case reflect.Struct:
			switch f := fType.Type.String(); f {
			case "time.Time":
				values[columnName] = fValue.Interface().(time.Time)
			case "null.Time":
				internalValue := fValue.Interface().(null.Time)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Time
				} else {
					values[columnName] = nil
				}
			case "null.String":
				internalValue := fValue.Interface().(null.String)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.String
				} else {
					values[columnName] = nil
				}
			case "null.JSON":
				internalValue := fValue.Interface().(null.JSON)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.JSON
				} else {
					values[columnName] = nil
				}
			case "null.Int":
				internalValue := fValue.Interface().(null.Int)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Int
				} else {
					values[columnName] = nil
				}
			case "null.Int8":
				internalValue := fValue.Interface().(null.Int8)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Int8
				} else {
					values[columnName] = nil
				}
			case "null.Int16":
				internalValue := fValue.Interface().(null.Int16)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Int16
				} else {
					values[columnName] = nil
				}
			case "null.Int32":
				internalValue := fValue.Interface().(null.Int32)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Int32
				} else {
					values[columnName] = nil
				}
			case "null.Int64":
				internalValue := fValue.Interface().(null.Int64)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Int64
				} else {
					values[columnName] = nil
				}
			case "null.Uint":
				internalValue := fValue.Interface().(null.Uint)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Uint
				} else {
					values[columnName] = nil
				}
			case "null.Uint8":
				internalValue := fValue.Interface().(null.Uint8)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Uint8
				} else {
					values[columnName] = nil
				}
			case "null.Uint16":
				internalValue := fValue.Interface().(null.Uint16)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Uint16
				} else {
					values[columnName] = nil
				}
			case "null.Uint32":
				internalValue := fValue.Interface().(null.Uint32)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Uint32
				} else {
					values[columnName] = nil
				}
			case "null.Uint64":
				internalValue := fValue.Interface().(null.Uint64)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Uint64
				} else {
					values[columnName] = nil
				}
			case "null.Bool":
				internalValue := fValue.Interface().(null.Bool)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Bool
				} else {
					values[columnName] = nil
				}
			case "null.Bytes":
				internalValue := fValue.Interface().(null.Bytes)
				if !internalValue.IsZero() {
					values[columnName] = internalValue.Bytes
				} else {
					values[columnName] = nil
				}
			case "null.Byte":
				internalValue := fValue.Interface().(null.Byte)
				if !internalValue.IsZero() {
					values[columnName] = int(internalValue.Byte)
				} else {
					values[columnName] = nil
				}
			default:
				internalDataset, err := RecognizeImitatorModel(fValue.Interface())
				if err != nil {
					return nil, fmt.Errorf("could not parse %s field: %w", columnName, err)
				}
				values[columnName] = internalDataset
			}
		default:
			values[columnName] = fValue.Interface()
		}

	}

	return &values, nil
}

type ImitatorModel map[string]interface{}

func (m *ImitatorModel) GetValue(table string, column string) (interface{}, bool) {
	raw := m

	if table != "" && table != "*" && table != "-" {
		exists := false
		if relation, ok := (*raw)["R"]; ok {
			var t *ImitatorModel
			if t, ok = relation.(*ImitatorModel); ok && t != nil {
				if relation, ok = (*t)[table]; ok && relation != nil {
					if t, ok = relation.(*ImitatorModel); ok && t != nil {
						raw = t
						exists = true
					}
				}
			}
		}
		if !exists {
			return nil, false
		}
	}

	value, ok := (*raw)[column]
	if !ok {
		return nil, false
	}

	return value, true
}

func (m *ImitatorModel) Compare(operator expression.Operator, table, column string, expected interface{}) (bool, error) {
	table = toLowerTableName(table)
	actually, exists := m.GetValue(table, column)
	if !exists {
		return false, fmt.Errorf(strings.Trim(fmt.Sprintf("%s.%s not found", table, column), "."))
	}

	switch operator {
	case expression.IsNull:
		return actually == nil, nil
	case expression.IsNotNull:
		return actually != nil, nil
	}

	switch val := actually.(type) {
	case string:
		expectedItems, err := asSlice[string](expected)
		if err != nil {
			return false, err
		}
		return compare(operator, val, expectedItems...)
	case int:
		expectedItems, err := asSlice[int](expected)
		if err != nil {
			return false, err
		}
		return compare(operator, val, expectedItems...)
	case uint:
		expectedItems, err := asSlice[uint](expected)
		if err != nil {
			return false, err
		}
		return compare(operator, val, expectedItems...)
	case float64:
		expectedItems, err := asSlice[float64](expected)
		if err != nil {
			return false, err
		}
		return compare(operator, val, expectedItems...)
	case time.Time:
		expectedItems, err := asSlice[time.Time](expected)
		if err != nil {
			return false, err
		}
		a := val.Format(time.RFC3339)
		e := make([]string, len(expectedItems))
		for i, t := range expectedItems {
			e[i] = t.UTC().Format(time.RFC3339)
		}
		return compare(operator, a, e...)
	case bool:
		expectedItems, err := asSlice[bool](expected)
		if err != nil {
			return false, err
		}
		a := 0
		if val {
			a = 1
		}
		e := make([]int, len(expectedItems))
		for i, t := range expectedItems {
			if t {
				e[i] = 1
			}
		}
		return compare(operator, a, e...)
	case []byte:
		expectedItems, err := asSlice[[]byte](expected)
		if err != nil {
			return false, err
		}
		a := hex.EncodeToString(val)
		e := make([]string, len(expectedItems))
		for i, t := range expectedItems {
			e[i] = hex.EncodeToString(t)
		}
		return compare(operator, a, e...)
	}

	return false, fmt.Errorf("unsupported type: %T %s %T", actually, operator, expected)
}

// TODO: rethink this function, it's not covered all cases
func toLowerTableName(str string) string {
	rx := regexp.MustCompile(`[^a-z0-9]+`)
	str = strings.ToLower(str)
	str = rx.ReplaceAllString(str, "")
	// reduce the plurality of the table name
	if strings.HasSuffix(str, "ies") {
		str = str[:len(str)-3] + "y"
	} else if strings.HasSuffix(str, "es") {
		str = str[:len(str)-1]
	} else if strings.HasSuffix(str, "s") {
		str = str[:len(str)-1]
	}
	return str
}

// compare compares a pair of value by the given operator.
func compare[T constraints.Ordered](o expression.Operator, actually T, expected ...T) (bool, error) {
	switch o {
	case expression.Equal:
		return actually == expected[0], nil
	case expression.NotEqual:
		return actually != expected[0], nil
	case expression.GreaterThan:
		return actually > expected[0], nil
	case expression.GreaterThanOrEqual:
		return actually >= expected[0], nil
	case expression.LessThan:
		return actually < expected[0], nil
	case expression.LessThanOrEqual:
		return actually <= expected[0], nil
	case expression.IsNull:
		var nilValue T
		return actually == nilValue, nil
	case expression.IsNotNull:
		var nilValue T
		return actually != nilValue, nil
	case expression.In:
		for _, e := range expected {
			if actually == e {
				return true, nil
			}
		}
		return false, nil
	case expression.NotIn:
		for _, e := range expected {
			if actually == e {
				return false, nil
			}
		}
		return true, nil
	case expression.Contains:
		return strings.Contains(fmt.Sprintf("%v", actually), fmt.Sprintf("%v", expected[0])), nil
	case expression.StartsWith:
		return strings.HasPrefix(fmt.Sprintf("%v", actually), fmt.Sprintf("%v", expected[0])), nil
	case expression.EndsWith:
		return strings.HasSuffix(fmt.Sprintf("%v", actually), fmt.Sprintf("%v", expected[0])), nil
	}
	return false, fmt.Errorf("unsupported operator: %T %s %T", actually, o, expected)
}

// asSlice converts the given value to a slice of the given type.
func asSlice[T any](v interface{}) ([]T, error) {
	if items, ok := v.([]interface{}); ok {
		var tmp []T
		for i, item := range items {
			var val T
			if val, ok = item.(T); !ok {
				return nil, fmt.Errorf("incorrect type: %T[%d]%T != %T", v, i, item, val)
			}
			tmp = append(tmp, val)
		}
		return tmp, nil
	}
	if val, ok := v.(T); ok {
		return []T{val}, nil
	}
	var val T
	return nil, fmt.Errorf("incorrect type: %T != %T", v, val)
}
