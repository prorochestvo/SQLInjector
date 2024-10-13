package schema

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"io"
	"net/http"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"
)

func NewInstruction(id, up, down string) Instruction {
	return Instruction{&instruction{id: id, up: up, down: down}}
}

// Instruction is a list of instructions
type Instruction []interface {
	ID() string
	MD5() string
}

// migration is a migration that is from database
type migration struct {
	id  string
	md5 string
}

func (m *migration) ID() string {
	return m.id
}

func (m *migration) MD5() string {
	return m.md5
}

// instruction is a migration that is not from database
type instruction struct {
	id   string
	up   string
	down string
}

func (i *instruction) ID() string {
	return i.id
}

func (i *instruction) Up() string {
	return i.up
}

func (i *instruction) Down() string {
	return i.down
}

func (i *instruction) MD5() string {
	hash := md5.New()
	hash.Write([]byte(i.up + "\n" + i.down))
	h := hash.Sum(nil)
	return hex.EncodeToString(h)
}

// MakeTableInstruction creates a migration instruction for SQLite3 database from a struct
func MakeTableInstruction(tableName string, tableFields interface{}, dialect internal.Dialect) (Instruction, error) {
	fields, err := parseTableFields(tableFields)
	if err != nil {
		return nil, err
	}

	f := make([]string, 0, len(fields))
	for k, v := range fields {
		var t string
		t, err = v.toSqlType(dialect)
		if err != nil {
			return nil, err
		}
		if k == "id" {
			t += " PRIMARY KEY"
		}
		f = append(f, k+" "+t)
	}

	up := "CREATE" + " TABLE IF NOT EXISTS " + tableName + " (\n   " + strings.Join(f, ",\n   ") + ");"
	down := "DROP" + " TABLE IF EXISTS " + tableName + ";"

	m := &instruction{
		id:   fmt.Sprintf("%d_%s_create", time.Now().UTC().Unix(), tableName),
		up:   up,
		down: down,
	}

	return Instruction{m}, nil
}

// ExtractInstructions extracts migrations from a folder
func ExtractInstructions(folder http.FileSystem, root string) (Instruction, error) {
	res := make(Instruction, 0)

	file, err := folder.Open(root)
	if err != nil {
		return nil, err
	}

	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		n := info.Name()
		n = strings.ToLower(n)
		if strings.HasSuffix(n, ".sql") {
			filePath := path.Join(root, info.Name())

			var f http.File
			f, err = folder.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("error while opening %s: %s", filePath, err.Error())
			}
			defer func(closer io.Closer) { err = errors.Join(err, closer.Close()) }(f)

			var i *instruction
			i, err = parseMigration(info.Name(), f)
			if err != nil {
				return nil, fmt.Errorf("error while parsing %s: %s", filePath, err.Error())
			}

			res = append(res, i)
		}
	}

	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID() < res[j].ID()
	})

	return res, nil
}

// parseMigration parses a migration from a file
func parseMigration(id string, r io.ReadSeeker) (*instruction, error) {
	i := &instruction{id: id, up: "", down: ""}

	var buf bytes.Buffer
	var accumulator *string

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 32*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// ignore comment except beginning with '--- #'
		if strings.Contains(line, "--") {
			trimLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimLine, MigrationCommandPrefix) {
				trimLine = strings.TrimPrefix(trimLine, MigrationCommandPrefix)
				trimLine = strings.TrimSpace(trimLine)
				if fields := strings.Fields(trimLine); len(fields) == 0 {
					return nil, fmt.Errorf(`ERROR: incomplete migration command`)
				} else {
					if accumulator != nil {
						if sqlScript := strings.TrimSpace(buf.String()); len(sqlScript) > 0 {
							*accumulator = sqlScript
						}
					}
					buf.Reset()
					switch d := strings.ToLower(fields[0]); d {
					case MigrationDirectionUp:
						accumulator = &i.up
					case MigrationDirectionDown:
						accumulator = &i.down
					default:
						accumulator = nil
					}
				}
				continue
			}
			if strings.HasPrefix(trimLine, "--") {
				continue
			}
		}

		if accumulator == nil {
			continue
		}

		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, err
		}
	}

	if sqlScript := strings.TrimSpace(buf.String()); len(sqlScript) > 0 && accumulator != nil {
		*accumulator = sqlScript
	}

	buf.Reset()

	return i, nil
}

// parseTableFields extracts table fields from a struct
func parseTableFields(o interface{}) (map[string]fieldType, error) {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type %s of object %T", t.Kind().String(), o)
	}

	fieldLength := t.NumField()
	tableFields := make(map[string]fieldType, fieldLength)

	for i := 0; i < fieldLength; i++ {
		fType := t.Field(i)
		fValue := v.Field(i)

		columnName := fType.Tag.Get("boil")
		columnName = strings.TrimSpace(columnName)
		if columnName == "" || columnName == "-" {
			continue
		}

		switch fKind := fValue.Kind(); fKind {
		case reflect.Pointer:
			tableFields[columnName] = fieldTypePointer
		case reflect.String:
			tableFields[columnName] = fieldTypeString
		case reflect.Int:
			tableFields[columnName] = fieldTypeInt
		case reflect.Uint:
			tableFields[columnName] = fieldTypeUint
		case reflect.Int8:
			tableFields[columnName] = fieldTypeInt8
		case reflect.Uint8:
			tableFields[columnName] = fieldTypeUint8
		case reflect.Int16:
			tableFields[columnName] = fieldTypeInt16
		case reflect.Uint16:
			tableFields[columnName] = fieldTypeUint16
		case reflect.Int32:
			tableFields[columnName] = fieldTypeInt32
		case reflect.Uint32:
			tableFields[columnName] = fieldTypeUint32
		case reflect.Int64:
			tableFields[columnName] = fieldTypeInt64
		case reflect.Uint64:
			tableFields[columnName] = fieldTypeUint64
		case reflect.Float32:
			tableFields[columnName] = fieldTypeFloat32
		case reflect.Float64:
			tableFields[columnName] = fieldTypeFloat64
		case reflect.Bool:
			tableFields[columnName] = fieldTypeBool
		case reflect.Slice, reflect.Array:
			println(fmt.Sprintf("<<< parseTableFields: %s | %s | %s", columnName, fValue.Type().Elem().Kind(), fType.Type.String()))
			switch fSliceKind := fValue.Type().Elem().Kind(); fSliceKind {
			case reflect.String:
				tableFields[columnName] = fieldTypeStringList
			case reflect.Bool:
				tableFields[columnName] = fieldTypeBoolList
			case reflect.Float64:
				tableFields[columnName] = fieldTypeFloat64List
			case reflect.Int64:
				tableFields[columnName] = fieldTypeInt64List
			default:
				switch fStructType := fType.Type.String(); fStructType {
				case "[]byte", "[]uint8":
					tableFields[columnName] = fieldTypeBytes
				case "null.JSON":
					tableFields[columnName] = fieldTypeNullJson
				case "types.JSON":
					tableFields[columnName] = fieldTypeJson
				case "types.DecimalArray":
					tableFields[columnName] = fieldTypeDecimalList
				default:
					if fSliceKind != reflect.Uint8 {
						return nil, fmt.Errorf("unsupported slice type %s of field %s | %s", fStructType, columnName, fSliceKind)
					}
					tableFields[columnName] = fieldTypeBytes
				}
			}
		case reflect.Struct:
			switch fStructType := fType.Type.String(); fStructType {
			case "time.Time":
				tableFields[columnName] = fieldTypeTime
			case "null.String":
				tableFields[columnName] = fieldTypeNullString
			case "null.Int":
				tableFields[columnName] = fieldTypeNullInt
			case "null.Uint":
				tableFields[columnName] = fieldTypeNullUint
			case "null.Int8":
				tableFields[columnName] = fieldTypeNullInt8
			case "null.Uint8":
				tableFields[columnName] = fieldTypeNullUint8
			case "null.Int16":
				tableFields[columnName] = fieldTypeNullInt16
			case "null.Uint16":
				tableFields[columnName] = fieldTypeNullUint16
			case "null.Int32":
				tableFields[columnName] = fieldTypeNullInt32
			case "null.Uint32":
				tableFields[columnName] = fieldTypeNullUint32
			case "null.Int64":
				tableFields[columnName] = fieldTypeNullInt64
			case "null.Uint64":
				tableFields[columnName] = fieldTypeNullUint64
			case "null.Float32":
				tableFields[columnName] = fieldTypeNullFloat32
			case "null.Float64":
				tableFields[columnName] = fieldTypeNullFloat64
			case "null.JSON":
				tableFields[columnName] = fieldTypeNullJson
			case "null.Bytes":
				tableFields[columnName] = fieldTypeNullBytes
			case "null.Bool":
				tableFields[columnName] = fieldTypeNullBool
			case "null.Time":
				tableFields[columnName] = fieldTypeNullTime
			case "types.JSON":
				tableFields[columnName] = fieldTypeJson
			case "types.Decimal":
				tableFields[columnName] = fieldTypeDecimal
			case "types.NullDecimal":
				tableFields[columnName] = fieldTypeNullDecimal
			case "pgeo.Point":
				tableFields[columnName] = fieldTypePoint
			case "pgeo.NullPoint":
				tableFields[columnName] = fieldTypeNullPoint
			default:
				return nil, fmt.Errorf("unsupported struct type %s of field %s", fStructType, columnName)
			}
		default:
			return nil, fmt.Errorf("unsupported type %s of field %s", fKind.String(), columnName)
		}
	}

	return tableFields, nil
}

const (
	MigrationCommandPrefix      = "--- #migrate:"
	MigrationDirectionUp        = "up"
	MigrationDirectionDown      = "down"
	migrationDirectionUndefined = ""

	fieldTypePointer     fieldType = "pointer"
	fieldTypeString      fieldType = "string"
	fieldTypeInt         fieldType = "int"
	fieldTypeUint        fieldType = "uint"
	fieldTypeInt8        fieldType = "int8"
	fieldTypeUint8       fieldType = "uint8"
	fieldTypeInt16       fieldType = "int16"
	fieldTypeUint16      fieldType = "uint16"
	fieldTypeInt32       fieldType = "int32"
	fieldTypeUint32      fieldType = "uint32"
	fieldTypeInt64       fieldType = "int64"
	fieldTypeUint64      fieldType = "uint64"
	fieldTypeFloat32     fieldType = "float32"
	fieldTypeFloat64     fieldType = "float64"
	fieldTypeJson        fieldType = "json"
	fieldTypeBytes       fieldType = "bytes"
	fieldTypeBool        fieldType = "bool"
	fieldTypeTime        fieldType = "time"
	fieldTypeNullString  fieldType = "null.string"
	fieldTypeNullInt     fieldType = "null.int"
	fieldTypeNullUint    fieldType = "null.uint"
	fieldTypeNullInt8    fieldType = "null.int8"
	fieldTypeNullUint8   fieldType = "null.uint8"
	fieldTypeNullInt16   fieldType = "null.int16"
	fieldTypeNullUint16  fieldType = "null.uint16"
	fieldTypeNullInt32   fieldType = "null.int32"
	fieldTypeNullUint32  fieldType = "null.uint32"
	fieldTypeNullInt64   fieldType = "null.int64"
	fieldTypeNullUint64  fieldType = "null.uint64"
	fieldTypeNullFloat32 fieldType = "null.float32"
	fieldTypeNullFloat64 fieldType = "null.float64"
	fieldTypeNullJson    fieldType = "null.json"
	fieldTypeNullBytes   fieldType = "null.bytes"
	fieldTypeNullBool    fieldType = "null.bool"
	fieldTypeNullTime    fieldType = "null.time"
	fieldTypeStringList  fieldType = "string_array"
	fieldTypeBoolList    fieldType = "bool_array"
	fieldTypeFloat64List fieldType = "float64_array"
	fieldTypeInt64List   fieldType = "int64_array"
	fieldTypeDecimalList fieldType = "decimal_array"
	fieldTypeDecimal     fieldType = "types.Decimal"
	fieldTypeNullDecimal fieldType = "types.NullDecimal"
	fieldTypePoint       fieldType = "pgeo.Point"
	fieldTypeNullPoint   fieldType = "pgeo.NullPoint"
)

type fieldType string

func (f fieldType) toSqlType(dialect internal.Dialect) (string, error) {
	switch dialect {
	case internal.DialectSQLite3:
		switch f {
		case fieldTypePointer:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeString:
			return "TEXT NOT NULL DEFAULT ''", nil
		case fieldTypeInt:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt8:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint8:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt16:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint16:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt32:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint32:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt64:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint64:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeFloat32:
			return "REAL NOT NULL DEFAULT 0.0", nil
		case fieldTypeFloat64:
			return "REAL NOT NULL DEFAULT 0.0", nil
		case fieldTypeJson:
			return "TEXT NOT NULL DEFAULT '{}'", nil
		case fieldTypeBytes:
			return "BLOB NOT NULL", nil
		case fieldTypeBool:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeTime:
			return "TEXT NOT NULL", nil
		case fieldTypeNullString:
			return "TEXT", nil
		case fieldTypeNullInt:
			return "INTEGER", nil
		case fieldTypeNullUint:
			return "INTEGER", nil
		case fieldTypeNullInt8:
			return "INTEGER", nil
		case fieldTypeNullUint8:
			return "INTEGER", nil
		case fieldTypeNullInt16:
			return "INTEGER", nil
		case fieldTypeNullUint16:
			return "INTEGER", nil
		case fieldTypeNullInt32:
			return "INTEGER", nil
		case fieldTypeNullUint32:
			return "INTEGER", nil
		case fieldTypeNullInt64:
			return "INTEGER", nil
		case fieldTypeNullUint64:
			return "INTEGER", nil
		case fieldTypeNullFloat32:
			return "REAL", nil
		case fieldTypeNullFloat64:
			return "REAL", nil
		case fieldTypeNullJson:
			return "TEXT", nil
		case fieldTypeNullBytes:
			return "BLOB", nil
		case fieldTypeNullBool:
			return "INTEGER", nil
		case fieldTypeNullTime:
			return "TEXT", nil
		case fieldTypeStringList:
			return "TEXT", nil
		case fieldTypeBoolList:
			return "TEXT", nil
		case fieldTypeFloat64List:
			return "TEXT", nil
		case fieldTypeInt64List:
			return "TEXT", nil
		case fieldTypeDecimalList:
			return "TEXT", nil
		case fieldTypeDecimal:
			return "TEXT", nil
		case fieldTypeNullDecimal:
			return "TEXT", nil
		case fieldTypePoint:
			return "TEXT", nil
		case fieldTypeNullPoint:
			return "TEXT", nil
		default:
			return "", fmt.Errorf("field type %s is not supported for %v", f, dialect)
		}
	case internal.DialectPostgreSQL:
		switch f {
		case fieldTypePointer:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeString:
			return "TEXT NOT NULL DEFAULT ''", nil
		case fieldTypeInt:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt8:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint8:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt16:
			return "SMALLINT NOT NULL DEFAULT 0", nil
		case fieldTypeUint16:
			return "SMALLINT NOT NULL DEFAULT 0", nil
		case fieldTypeInt32:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeUint32:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeInt64:
			return "BIGINT NOT NULL DEFAULT 0", nil
		case fieldTypeUint64:
			return "BIGINT NOT NULL DEFAULT 0", nil
		case fieldTypeFloat32:
			return "REAL NOT NULL DEFAULT 0.0", nil
		case fieldTypeFloat64:
			return "DOUBLE PRECISION NOT NULL DEFAULT 0.0", nil
		case fieldTypeJson:
			return "JSONB NOT NULL DEFAULT '{}'::JSONB", nil
		case fieldTypeBool:
			return "BOOLEAN NOT NULL DEFAULT FALSE", nil
		case fieldTypeTime:
			return "TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP", nil
		case fieldTypeBytes:
			return "BLOB NOT NULL DEFAULT ''", nil
		case fieldTypeNullString:
			return "TEXT", nil
		case fieldTypeNullInt:
			return "INTEGER", nil
		case fieldTypeNullUint:
			return "INTEGER", nil
		case fieldTypeNullInt8:
			return "INTEGER", nil
		case fieldTypeNullUint8:
			return "INTEGER", nil
		case fieldTypeNullInt16:
			return "SMALLINT", nil
		case fieldTypeNullUint16:
			return "SMALLINT", nil
		case fieldTypeNullInt32:
			return "INTEGER", nil
		case fieldTypeNullUint32:
			return "INTEGER", nil
		case fieldTypeNullInt64:
			return "BIGINT", nil
		case fieldTypeNullUint64:
			return "BIGINT", nil
		case fieldTypeNullFloat32:
			return "REAL", nil
		case fieldTypeNullFloat64:
			return "DOUBLE PRECISION", nil
		case fieldTypeNullJson:
			return "JSONB", nil
		case fieldTypeNullBytes:
			return "SEIL", nil
		case fieldTypeNullBool:
			return "BOOLEAN", nil
		case fieldTypeNullTime:
			return "TIMESTAMPTZ", nil
		case fieldTypeStringList:
			return "TEXT[] NOT NULL DEFAULT '{}'::TEXT[]", nil
		case fieldTypeBoolList:
			return "BOOLEAN[] NOT NULL DEFAULT '{}'::BOOLEAN[]", nil
		case fieldTypeFloat64List:
			return "REAL[] NOT NULL DEFAULT '{}'::REAL[]", nil
		case fieldTypeInt64List:
			return "BIGINT[] NOT NULL DEFAULT '{}'::BIGINT[]", nil
		case fieldTypeDecimalList:
			return "DECIMAL[] NOT NULL DEFAULT '{}'::DECIMAL[]", nil
		case fieldTypeDecimal:
			return "DECIMAL NOT NULL DEFAULT 0", nil
		case fieldTypeNullDecimal:
			return "DECIMAL", nil
		case fieldTypePoint:
			return "INTEGER NOT NULL DEFAULT 0", nil
		case fieldTypeNullPoint:
			return "INTEGER", nil
		default:
			return "", fmt.Errorf("field type %s is not supported for %v", f, dialect)
		}
	}
	return "", fmt.Errorf("dialect %s is not supported", dialect)
}
