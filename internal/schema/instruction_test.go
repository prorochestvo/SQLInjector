package schema

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

var _ instructionDown = &instruction{}
var _ instructionDown = &instruction{}

func TestExtractInstructions(t *testing.T) {
	folder := t.TempDir()
	t.Run("OS.Files", func(t *testing.T) {
		tableName := "os_files_01"
		createTable := fmt.Sprintf("%sUp\nCREATE TABLE %s (id TEXT PRIMARY KEY);\n\n%sDown\nDROP TABLE %s;\n\n", MigrationCommandPrefix, tableName, MigrationCommandPrefix, tableName)
		insertTable01 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", MigrationCommandPrefix, tableName, uuid.NewV4().String(), MigrationCommandPrefix, tableName)
		insertTable02 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", MigrationCommandPrefix, tableName, uuid.NewV4().String(), MigrationCommandPrefix, tableName)
		insertTable03 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", MigrationCommandPrefix, tableName, uuid.NewV4().String(), MigrationCommandPrefix, tableName)
		require.NoError(t, os.WriteFile(path.Join(folder, "003.insert_table_02.sql"), []byte(insertTable02), 0666))
		require.NoError(t, os.WriteFile(path.Join(folder, "001.create_table.sql"), []byte(createTable), 0666))
		require.NoError(t, os.WriteFile(path.Join(folder, "001.create_table.txt"), []byte(createTable), 0666))
		require.NoError(t, os.WriteFile(path.Join(folder, "002.insert_table_01.sql"), []byte(insertTable01), 0666))
		require.NoError(t, os.WriteFile(path.Join(folder, "004.insert_table_03.sql"), []byte(insertTable03), 0666))

		filesystem := http.Dir(folder)
		migrations, err := ExtractInstructions(filesystem, "/")
		require.NoError(t, err)
		require.NotNil(t, migrations)

		expectedValues := []string{createTable, insertTable01, insertTable02, insertTable03}
		require.Len(t, migrations, len(expectedValues))
		for i, expected := range expectedValues {
			m := migrations[i]
			actually := fmt.Sprintf("%sUp\n%s\n\n%sDown\n%s\n\n", MigrationCommandPrefix, m.(instructionUp).Up(), MigrationCommandPrefix, m.(instructionDown).Down())
			require.Equal(t, expected, actually)
		}
	})
	t.Run("FS.Files", func(t *testing.T) {
		filesystem := http.FS(internalMigrations)
		migrations, err := ExtractInstructions(filesystem, "/")
		require.NoError(t, err)
		require.NotNil(t, migrations)

		require.Len(t, migrations, 2)
		require.Equal(t, "CREATE "+"TABLE os_files_01 (id TEXT PRIMARY KEY);", migrations[0].(instructionUp).Up())
		require.Equal(t, "DROP "+"TABLE os_files_01;", migrations[0].(instructionDown).Down())
		require.Equal(t, "INSERT "+"INTO os_files_01 (id) VALUES ('bd5dc0fa-db1a-4e15-bea4-34c4fcc2133b');", migrations[1].(instructionUp).Up())
		require.Equal(t, "DELETE "+"FROM os_files_01;", migrations[1].(instructionDown).Down())
	})
}

func TestParseMigration(t *testing.T) {
	t.Run("UpAndDown", func(t *testing.T) {
		mUp01 := `CREATE EXTENSION IF NOT EXISTS "uuid-osSP";`
		mUp02 := `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`
		mDown01 := `DROP EXTENSION IF EXISTS "uuid-osSP";`
		mDown02 := `DROP EXTENSION IF EXISTS "pgcrypto";`
		n := "M001"
		m, err := parseMigration(n, bytes.NewReader([]byte(" "+MigrationCommandPrefix+" up\n"+mUp01+"\n\n"+mUp02+"\n\n"+MigrationCommandPrefix+" down\n"+mDown01+"\n\n"+mDown02+"\n\n")))
		require.NoError(t, err)
		require.NotNil(t, m)
		require.Equal(t, n, m.id)
		require.Equal(t, mUp01+"\n\n"+mUp02, m.up)
		require.Equal(t, mDown01+"\n\n"+mDown02, m.down)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Skip("not implemented")
	})
	t.Run("UpEmptyOnly", func(t *testing.T) {
		t.Skip("not implemented")
	})
	t.Run("DownEmptyOnly", func(t *testing.T) {
		t.Skip("not implemented")
	})
}

func TestParseTableFields(t *testing.T) {
	t.Run("*User", func(t *testing.T) {
		tableFields, err := parseTableFields(&internalUser{})
		require.NoError(t, err)
		require.NotNil(t, tableFields)
		require.Len(t, tableFields, 10)

		require.Equal(t, fieldTypeString, tableFields["id"])
		require.Equal(t, fieldTypeString, tableFields["login"])
		require.Equal(t, fieldTypeNullString, tableFields["password"])
		require.Equal(t, fieldTypeNullString, tableFields["nickname"])
		require.Equal(t, fieldTypeNullJson, tableFields["personal_dataset"])
		require.Equal(t, fieldTypeString, tableFields["default_local"])
		require.Equal(t, fieldTypeInt, tableFields["default_role"])
		require.Equal(t, fieldTypeTime, tableFields["created_at"])
		require.Equal(t, fieldTypeTime, tableFields["updated_at"])
		require.Equal(t, fieldTypeNullTime, tableFields["deleted_at"])
	})
	t.Run("User", func(t *testing.T) {
		tableFields, err := parseTableFields(internalUser{})
		require.NoError(t, err)
		require.NotNil(t, tableFields)
		require.Len(t, tableFields, 10)

		require.Equal(t, fieldTypeString, tableFields["id"])
		require.Equal(t, fieldTypeString, tableFields["login"])
		require.Equal(t, fieldTypeNullString, tableFields["password"])
		require.Equal(t, fieldTypeNullString, tableFields["nickname"])
		require.Equal(t, fieldTypeNullJson, tableFields["personal_dataset"])
		require.Equal(t, fieldTypeString, tableFields["default_local"])
		require.Equal(t, fieldTypeInt, tableFields["default_role"])
		require.Equal(t, fieldTypeTime, tableFields["created_at"])
		require.Equal(t, fieldTypeTime, tableFields["updated_at"])
		require.Equal(t, fieldTypeNullTime, tableFields["deleted_at"])
	})
	t.Run("Types", func(t *testing.T) {
		o := struct {
			FString            string             `boil:"f_string"`
			FInt               int                `boil:"f_int"`
			FUint              uint               `boil:"f_uint"`
			FInt8              int8               `boil:"f_int8"`
			FUint8             uint8              `boil:"f_uint8"`
			FInt16             int16              `boil:"f_int16"`
			FUint16            uint16             `boil:"f_uint16"`
			FInt32             int32              `boil:"f_int32"`
			FUint32            uint32             `boil:"f_uint32"`
			FInt64             int64              `boil:"f_int64"`
			FUint64            uint64             `boil:"f_uint64"`
			FFloat32           float32            `boil:"f_float32"`
			FFloat64           float64            `boil:"f_float64"`
			FBytes             []byte             `boil:"f_bytes"`
			FBool              bool               `boil:"f_bool"`
			FTimeTime          time.Time          `boil:"f_timeTime"`
			FNullString        null.String        `boil:"f_null_string"`
			FNullInt           null.Int           `boil:"f_null_int"`
			FNullUint          null.Uint          `boil:"f_null_uint"`
			FNullInt8          null.Int8          `boil:"f_null_int8"`
			FNullUint8         null.Uint8         `boil:"f_null_uint8"`
			FNullInt16         null.Int16         `boil:"f_null_int16"`
			FNullUint16        null.Uint16        `boil:"f_null_uint16"`
			FNullInt32         null.Int32         `boil:"f_null_int32"`
			FNullUint32        null.Uint32        `boil:"f_null_uint32"`
			FNullInt64         null.Int64         `boil:"f_null_int64"`
			FNullUint64        null.Uint64        `boil:"f_null_uint64"`
			FNullFloat32       null.Float32       `boil:"f_null_float32"`
			FNullFloat64       null.Float64       `boil:"f_null_float64"`
			FNullJSON          null.JSON          `boil:"f_null_json"`
			FNullBytes         null.Bytes         `boil:"f_null_bytes"`
			FNullBool          null.Bool          `boil:"f_null_bool"`
			FNullTime          null.Time          `boil:"f_null_time"`
			FTypesStringArray  types.StringArray  `boil:"f_types_string_array"`
			FTypesJSON         types.JSON         `boil:"f_types_json"`
			FTypesBoolArray    types.BoolArray    `boil:"f_types_bool_array"`
			FTypesFloat64Array types.Float64Array `boil:"f_types_float64_array"`
			FTypesInt64Array   types.Int64Array   `boil:"f_types_int64_array"`
			FTypesDecimalArray types.DecimalArray `boil:"f_types_decimal_array"`
			FTypesByte         types.Byte         `boil:"f_types_byte"`
			FTypesDecimal      types.Decimal      `boil:"f_types_decimal"`
			FTypesNullDecimal  types.NullDecimal  `boil:"f_types_null_decimal"`
			FPgeoPoint         pgeo.Point         `boil:"f_pgeo_point"`
			FPgeoNullPoint     pgeo.NullPoint     `boil:"f_pgeo_null_point"`
		}{}
		tableFields, err := parseTableFields(o)
		require.NoError(t, err)
		require.NotNil(t, tableFields)
		require.Len(t, tableFields, 44)

		require.Equal(t, fieldTypeString, tableFields["f_string"])
		require.Equal(t, fieldTypeInt, tableFields["f_int"])
		require.Equal(t, fieldTypeUint, tableFields["f_uint"])
		require.Equal(t, fieldTypeInt8, tableFields["f_int8"])
		require.Equal(t, fieldTypeUint8, tableFields["f_uint8"])
		require.Equal(t, fieldTypeInt16, tableFields["f_int16"])
		require.Equal(t, fieldTypeUint16, tableFields["f_uint16"])
		require.Equal(t, fieldTypeInt32, tableFields["f_int32"])
		require.Equal(t, fieldTypeUint32, tableFields["f_uint32"])
		require.Equal(t, fieldTypeInt64, tableFields["f_int64"])
		require.Equal(t, fieldTypeUint64, tableFields["f_uint64"])
		require.Equal(t, fieldTypeFloat32, tableFields["f_float32"])
		require.Equal(t, fieldTypeFloat64, tableFields["f_float64"])
		require.Equal(t, fieldTypeBytes, tableFields["f_bytes"])
		require.Equal(t, fieldTypeBool, tableFields["f_bool"])
		require.Equal(t, fieldTypeTime, tableFields["f_timeTime"])
		require.Equal(t, fieldTypeNullString, tableFields["f_null_string"])
		require.Equal(t, fieldTypeNullInt, tableFields["f_null_int"])
		require.Equal(t, fieldTypeNullUint, tableFields["f_null_uint"])
		require.Equal(t, fieldTypeNullInt8, tableFields["f_null_int8"])
		require.Equal(t, fieldTypeNullUint8, tableFields["f_null_uint8"])
		require.Equal(t, fieldTypeNullInt16, tableFields["f_null_int16"])
		require.Equal(t, fieldTypeNullUint16, tableFields["f_null_uint16"])
		require.Equal(t, fieldTypeNullInt32, tableFields["f_null_int32"])
		require.Equal(t, fieldTypeNullUint32, tableFields["f_null_uint32"])
		require.Equal(t, fieldTypeNullInt64, tableFields["f_null_int64"])
		require.Equal(t, fieldTypeNullUint64, tableFields["f_null_uint64"])
		require.Equal(t, fieldTypeNullFloat32, tableFields["f_null_float32"])
		require.Equal(t, fieldTypeNullFloat64, tableFields["f_null_float64"])
		require.Equal(t, fieldTypeNullJson, tableFields["f_null_json"])
		require.Equal(t, fieldTypeNullBytes, tableFields["f_null_bytes"])
		require.Equal(t, fieldTypeNullBool, tableFields["f_null_bool"])
		require.Equal(t, fieldTypeNullTime, tableFields["f_null_time"])
		require.Equal(t, fieldTypeJson, tableFields["f_types_json"])
		require.Equal(t, fieldTypeStringList, tableFields["f_types_string_array"])
		require.Equal(t, fieldTypeBoolList, tableFields["f_types_bool_array"])
		require.Equal(t, fieldTypeFloat64List, tableFields["f_types_float64_array"])
		require.Equal(t, fieldTypeInt64List, tableFields["f_types_int64_array"])
		require.Equal(t, fieldTypeDecimalList, tableFields["f_types_decimal_array"])
		require.Equal(t, fieldTypeUint8, tableFields["f_types_byte"])
		require.Equal(t, fieldTypeDecimal, tableFields["f_types_decimal"])
		require.Equal(t, fieldTypeNullDecimal, tableFields["f_types_null_decimal"])
		require.Equal(t, fieldTypePoint, tableFields["f_pgeo_point"])
		require.Equal(t, fieldTypeNullPoint, tableFields["f_pgeo_null_point"])
	})
	t.Run("ForSQLite3", func(t *testing.T) {
		fields := map[string]fieldType{
			string(fieldTypePointer):     fieldTypePointer,
			string(fieldTypeString):      fieldTypeString,
			string(fieldTypeInt):         fieldTypeInt,
			string(fieldTypeUint):        fieldTypeUint,
			string(fieldTypeInt8):        fieldTypeInt8,
			string(fieldTypeUint8):       fieldTypeUint8,
			string(fieldTypeInt16):       fieldTypeInt16,
			string(fieldTypeUint16):      fieldTypeUint16,
			string(fieldTypeInt32):       fieldTypeInt32,
			string(fieldTypeUint32):      fieldTypeUint32,
			string(fieldTypeInt64):       fieldTypeInt64,
			string(fieldTypeUint64):      fieldTypeUint64,
			string(fieldTypeFloat32):     fieldTypeFloat32,
			string(fieldTypeFloat64):     fieldTypeFloat64,
			string(fieldTypeJson):        fieldTypeJson,
			string(fieldTypeBytes):       fieldTypeBytes,
			string(fieldTypeBool):        fieldTypeBool,
			string(fieldTypeTime):        fieldTypeTime,
			string(fieldTypeNullString):  fieldTypeNullString,
			string(fieldTypeNullInt):     fieldTypeNullInt,
			string(fieldTypeNullUint):    fieldTypeNullUint,
			string(fieldTypeNullInt8):    fieldTypeNullInt8,
			string(fieldTypeNullUint8):   fieldTypeNullUint8,
			string(fieldTypeNullInt16):   fieldTypeNullInt16,
			string(fieldTypeNullUint16):  fieldTypeNullUint16,
			string(fieldTypeNullInt32):   fieldTypeNullInt32,
			string(fieldTypeNullUint32):  fieldTypeNullUint32,
			string(fieldTypeNullInt64):   fieldTypeNullInt64,
			string(fieldTypeNullUint64):  fieldTypeNullUint64,
			string(fieldTypeNullFloat32): fieldTypeNullFloat32,
			string(fieldTypeNullFloat64): fieldTypeNullFloat64,
			string(fieldTypeNullJson):    fieldTypeNullJson,
			string(fieldTypeNullBytes):   fieldTypeNullBytes,
			string(fieldTypeNullBool):    fieldTypeNullBool,
			string(fieldTypeNullTime):    fieldTypeNullTime,
			string(fieldTypeStringList):  fieldTypeStringList,
			string(fieldTypeBoolList):    fieldTypeBoolList,
			string(fieldTypeFloat64List): fieldTypeFloat64List,
			string(fieldTypeInt64List):   fieldTypeInt64List,
			string(fieldTypeDecimalList): fieldTypeDecimalList,
			string(fieldTypeDecimal):     fieldTypeDecimal,
			string(fieldTypeNullDecimal): fieldTypeNullDecimal,
			string(fieldTypePoint):       fieldTypePoint,
			string(fieldTypeNullPoint):   fieldTypeNullPoint,
		}
		for fName, fType := range fields {
			fTypeSQL, err := fType.toSqlType(internal.DialectSQLite3)
			require.NoError(t, err, fName)
			require.NotNil(t, fTypeSQL, fName)
			t.Skip("not implemented for", fName)
		}
	})
	t.Run("ForPostgreSQL", func(t *testing.T) {
		fields := map[string]fieldType{
			string(fieldTypePointer):     fieldTypePointer,
			string(fieldTypeString):      fieldTypeString,
			string(fieldTypeInt):         fieldTypeInt,
			string(fieldTypeUint):        fieldTypeUint,
			string(fieldTypeInt8):        fieldTypeInt8,
			string(fieldTypeUint8):       fieldTypeUint8,
			string(fieldTypeInt16):       fieldTypeInt16,
			string(fieldTypeUint16):      fieldTypeUint16,
			string(fieldTypeInt32):       fieldTypeInt32,
			string(fieldTypeUint32):      fieldTypeUint32,
			string(fieldTypeInt64):       fieldTypeInt64,
			string(fieldTypeUint64):      fieldTypeUint64,
			string(fieldTypeFloat32):     fieldTypeFloat32,
			string(fieldTypeFloat64):     fieldTypeFloat64,
			string(fieldTypeJson):        fieldTypeJson,
			string(fieldTypeBytes):       fieldTypeBytes,
			string(fieldTypeBool):        fieldTypeBool,
			string(fieldTypeTime):        fieldTypeTime,
			string(fieldTypeNullString):  fieldTypeNullString,
			string(fieldTypeNullInt):     fieldTypeNullInt,
			string(fieldTypeNullUint):    fieldTypeNullUint,
			string(fieldTypeNullInt8):    fieldTypeNullInt8,
			string(fieldTypeNullUint8):   fieldTypeNullUint8,
			string(fieldTypeNullInt16):   fieldTypeNullInt16,
			string(fieldTypeNullUint16):  fieldTypeNullUint16,
			string(fieldTypeNullInt32):   fieldTypeNullInt32,
			string(fieldTypeNullUint32):  fieldTypeNullUint32,
			string(fieldTypeNullInt64):   fieldTypeNullInt64,
			string(fieldTypeNullUint64):  fieldTypeNullUint64,
			string(fieldTypeNullFloat32): fieldTypeNullFloat32,
			string(fieldTypeNullFloat64): fieldTypeNullFloat64,
			string(fieldTypeNullJson):    fieldTypeNullJson,
			string(fieldTypeNullBytes):   fieldTypeNullBytes,
			string(fieldTypeNullBool):    fieldTypeNullBool,
			string(fieldTypeNullTime):    fieldTypeNullTime,
			string(fieldTypeStringList):  fieldTypeStringList,
			string(fieldTypeBoolList):    fieldTypeBoolList,
			string(fieldTypeFloat64List): fieldTypeFloat64List,
			string(fieldTypeInt64List):   fieldTypeInt64List,
			string(fieldTypeDecimalList): fieldTypeDecimalList,
			string(fieldTypeDecimal):     fieldTypeDecimal,
			string(fieldTypeNullDecimal): fieldTypeNullDecimal,
			string(fieldTypePoint):       fieldTypePoint,
			string(fieldTypeNullPoint):   fieldTypeNullPoint,
		}
		for fName, fType := range fields {
			fTypeSQL, err := fType.toSqlType(internal.DialectPostgreSQL)
			require.NoError(t, err, fName)
			require.NotNil(t, fTypeSQL, fName)
			t.Skip("not implemented for", fName)
		}
	})
	t.Run("FieldsForMySQL", func(t *testing.T) {
		fields := map[string]fieldType{
			string(fieldTypePointer):     fieldTypePointer,
			string(fieldTypeString):      fieldTypeString,
			string(fieldTypeInt):         fieldTypeInt,
			string(fieldTypeUint):        fieldTypeUint,
			string(fieldTypeInt8):        fieldTypeInt8,
			string(fieldTypeUint8):       fieldTypeUint8,
			string(fieldTypeInt16):       fieldTypeInt16,
			string(fieldTypeUint16):      fieldTypeUint16,
			string(fieldTypeInt32):       fieldTypeInt32,
			string(fieldTypeUint32):      fieldTypeUint32,
			string(fieldTypeInt64):       fieldTypeInt64,
			string(fieldTypeUint64):      fieldTypeUint64,
			string(fieldTypeFloat32):     fieldTypeFloat32,
			string(fieldTypeFloat64):     fieldTypeFloat64,
			string(fieldTypeJson):        fieldTypeJson,
			string(fieldTypeBytes):       fieldTypeBytes,
			string(fieldTypeBool):        fieldTypeBool,
			string(fieldTypeTime):        fieldTypeTime,
			string(fieldTypeNullString):  fieldTypeNullString,
			string(fieldTypeNullInt):     fieldTypeNullInt,
			string(fieldTypeNullUint):    fieldTypeNullUint,
			string(fieldTypeNullInt8):    fieldTypeNullInt8,
			string(fieldTypeNullUint8):   fieldTypeNullUint8,
			string(fieldTypeNullInt16):   fieldTypeNullInt16,
			string(fieldTypeNullUint16):  fieldTypeNullUint16,
			string(fieldTypeNullInt32):   fieldTypeNullInt32,
			string(fieldTypeNullUint32):  fieldTypeNullUint32,
			string(fieldTypeNullInt64):   fieldTypeNullInt64,
			string(fieldTypeNullUint64):  fieldTypeNullUint64,
			string(fieldTypeNullFloat32): fieldTypeNullFloat32,
			string(fieldTypeNullFloat64): fieldTypeNullFloat64,
			string(fieldTypeNullJson):    fieldTypeNullJson,
			string(fieldTypeNullBytes):   fieldTypeNullBytes,
			string(fieldTypeNullBool):    fieldTypeNullBool,
			string(fieldTypeNullTime):    fieldTypeNullTime,
			string(fieldTypeStringList):  fieldTypeStringList,
			string(fieldTypeBoolList):    fieldTypeBoolList,
			string(fieldTypeFloat64List): fieldTypeFloat64List,
			string(fieldTypeInt64List):   fieldTypeInt64List,
			string(fieldTypeDecimalList): fieldTypeDecimalList,
			string(fieldTypeDecimal):     fieldTypeDecimal,
			string(fieldTypeNullDecimal): fieldTypeNullDecimal,
			string(fieldTypePoint):       fieldTypePoint,
			string(fieldTypeNullPoint):   fieldTypeNullPoint,
		}
		for fName, fType := range fields {
			_, err := fType.toSqlType(internal.DialectMySQL)
			require.Error(t, err, fName)
		}
	})
}

//go:embed instruction_test.*.sql
var internalMigrations embed.FS

type internalUser struct {
	ID              string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	Login           string      `boil:"login" json:"login" toml:"login" yaml:"login"`
	Password        null.String `boil:"password" json:"password,omitempty" toml:"password" yaml:"password,omitempty"`
	Nickname        null.String `boil:"nickname" json:"nickname,omitempty" toml:"nickname" yaml:"nickname,omitempty"`
	PersonalDataset null.JSON   `boil:"personal_dataset" json:"personal_dataset,omitempty" toml:"personal_dataset" yaml:"personal_dataset,omitempty"`
	DefaultLocal    string      `boil:"default_local" json:"default_local" toml:"default_local" yaml:"default_local"`
	DefaultRole     int         `boil:"default_role" json:"default_role" toml:"default_role" yaml:"default_role"`
	CreatedAt       time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	DeletedAt       null.Time   `boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *internalUserR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L internalUserL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

type internalUserR struct {
}

type internalUserL struct{}

type instructionDown interface {
	Down() string
}

type instructionUp interface {
	Up() string
}
