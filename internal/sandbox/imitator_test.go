package sandbox

import (
	"github.com/myesui/uuid"
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	m := &internalTask{ID: 1, Name: "N001", IsEnabled: false, SubjectID: 1}

	err := Merge(m, map[string]interface{}{
		"id":           int64(2),
		"name":         "N002",
		"subject_id":   int64(20),
		"subject_name": "math",
	})

	require.NoError(t, err)
	require.Equal(t, int64(2), m.ID)
	require.Equal(t, "N002", m.Name)
	require.Equal(t, int64(20), m.SubjectID)
}

func TestSimulate(t *testing.T) {
	t1 := &internalTask{ID: 1, Name: "N001", IsEnabled: false, SubjectID: 1}
	t2 := &internalTask{ID: 2, Name: "Num2", IsEnabled: false, SubjectID: 2}
	t3 := &internalTask{ID: 3, Name: "N003", IsEnabled: true, SubjectID: 3}
	t4 := &internalTask{ID: 4, Name: "num4", IsEnabled: false, SubjectID: 4}
	t5 := &internalTask{ID: 5, Name: "N005", IsEnabled: false, SubjectID: 5}
	t6 := &internalTask{ID: 6, Name: "Num6", IsEnabled: true, SubjectID: 1}
	t7 := &internalTask{ID: 7, Name: "N007", IsEnabled: false, SubjectID: 1}

	items := map[int64]*internalTask{
		t1.ID: t1,
		t2.ID: t2,
		t3.ID: t3,
		t4.ID: t4,
		t5.ID: t5,
		t6.ID: t6,
		t7.ID: t7,
	}
	t.Run("WHERE is_enabled = true", func(t *testing.T) {
		actually, err := ImitatorSql[int64, internalTask](
			items,
			[]*expression.Where{expression.NewWhere("is_enabled", expression.Equal, true)},
			nil,
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, actually)
		sort.SliceStable(actually, func(a, b int) bool {
			return actually[a].ID < actually[b].ID
		})
		require.Len(t, actually, 2)
		require.Equal(t, t3, actually[0])
		require.Equal(t, t6, actually[1])
	})
	t.Run("GROUP BY is_enabled", func(t *testing.T) {
		actually, err := ImitatorSql[int64, internalTask](
			items,
			nil,
			[]*expression.GroupBy{expression.NewGroupBy("is_enabled")},
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, actually)
		sort.SliceStable(actually, func(a, b int) bool {
			return actually[a].IsEnabled == true && actually[b].IsEnabled == false
		})
		require.Len(t, actually, 2)
		require.Equal(t, t6.IsEnabled, actually[0].IsEnabled)
		require.Equal(t, t7.IsEnabled, actually[1].IsEnabled)
	})
	t.Run("ORDER BY id", func(t *testing.T) {
		actually, err := ImitatorSql[int64, internalTask](
			items,
			nil,
			nil,
			[]*expression.OrderBy{expression.NewOrderBy("id", expression.Descending)},
		)
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 7)
		require.Equal(t, t7, actually[0])
		require.Equal(t, t6, actually[1])
		require.Equal(t, t5, actually[2])
		require.Equal(t, t4, actually[3])
		require.Equal(t, t3, actually[4])
		require.Equal(t, t2, actually[5])
		require.Equal(t, t1, actually[6])
	})
	t.Run("WHERE id in (6,7); GROUP BY is_enabled; ORDER BY id", func(t *testing.T) {
		actually, err := ImitatorSql[int64, internalTask](
			items,
			[]*expression.Where{expression.NewWhere("id", expression.In, 6, 3, 2, 7)},
			[]*expression.GroupBy{expression.NewGroupBy("subject_id")},
			[]*expression.OrderBy{expression.NewOrderBy("id", expression.Descending)},
		)
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 3)
		require.True(t, actually[0] == t7 || actually[0] == t6)
		require.True(t, actually[1] == t3)
		require.True(t, actually[2] == t2)
	})
}

func TestImitatorSqlWhere(t *testing.T) {
	t1 := &internalTask{ID: 1, Name: "N001", SubjectID: 1, IsEnabled: false, LastSyncError: null.StringFrom(""), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t2 := &internalTask{ID: 2, Name: "Num2", SubjectID: 2, IsEnabled: false, LastSyncError: null.StringFrom(""), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: true}}}
	t3 := &internalTask{ID: 3, Name: "N003", SubjectID: 1, IsEnabled: true, LastSyncError: null.StringFrom(""), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t4 := &internalTask{ID: 4, Name: "num4", SubjectID: 2, IsEnabled: false, LastSyncError: null.String{Valid: false}, DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t5 := &internalTask{ID: 5, Name: "N005", SubjectID: 1, IsEnabled: false, LastSyncError: null.StringFrom(""), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: true}}}
	t6 := &internalTask{ID: 6, Name: "Num6", SubjectID: 2, IsEnabled: true, LastSyncError: null.StringFrom(""), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t7 := &internalTask{ID: 7, Name: "N007", SubjectID: 1, IsEnabled: false, LastSyncError: null.StringFrom(""), DeletedAt: null.TimeFrom(time.Now().UTC()), R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}

	m1, err := RecognizeImitatorModel(t1)
	require.NoError(t, err)
	m2, err := RecognizeImitatorModel(t2)
	require.NoError(t, err)
	m3, err := RecognizeImitatorModel(t3)
	require.NoError(t, err)
	m4, err := RecognizeImitatorModel(t4)
	require.NoError(t, err)
	m5, err := RecognizeImitatorModel(t5)
	require.NoError(t, err)
	m6, err := RecognizeImitatorModel(t6)
	require.NoError(t, err)
	m7, err := RecognizeImitatorModel(t7)
	require.NoError(t, err)

	items := []*ImitatorModel{
		m1,
		m2,
		m3,
		m4,
		m5,
		m6,
		m7,
	}

	t.Run("id eq 2", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("id", expression.Equal, 2))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 1)
		require.Equal(t, m2, actually[0])
	})
	t.Run("id in (2,3)", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("id", expression.In, 2, 3))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		require.Equal(t, m2, actually[0])
		require.Equal(t, m3, actually[1])
	})
	t.Run("isEnabled ne false", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("is_enabled", expression.NotEqual, false))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t3.Name, a[0])
		require.Equal(t, t6.Name, a[1])
	})
	t.Run("name contains Num", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("name", expression.Contains, "Num"))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t2.Name, a[0])
		require.Equal(t, t6.Name, a[1])
	})
	t.Run("subjectID eq 1 AND isEnabled eq true", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("subject_id", expression.Equal, 1), expression.NewWhere("is_enabled", expression.Equal, true))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 1)
		require.Equal(t, m3, actually[0])
	})
	t.Run("subjectID lt 2", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("subject_id", expression.LessThan, 2))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 4)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
			(*actually[2])["name"].(string),
			(*actually[3])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t1.Name, a[0])
		require.Equal(t, t3.Name, a[1])
		require.Equal(t, t5.Name, a[2])
		require.Equal(t, t7.Name, a[3])
	})
	t.Run("subjectID le 2", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("subject_id", expression.LessThanOrEqual, 2))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 7)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
			(*actually[2])["name"].(string),
			(*actually[3])["name"].(string),
			(*actually[4])["name"].(string),
			(*actually[5])["name"].(string),
			(*actually[6])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t1.Name, a[0])
		require.Equal(t, t3.Name, a[1])
		require.Equal(t, t5.Name, a[2])
		require.Equal(t, t7.Name, a[3])
		require.Equal(t, t2.Name, a[4])
		require.Equal(t, t6.Name, a[5])
		require.Equal(t, t4.Name, a[6])
	})
	t.Run("subjectID gt 1", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("subject_id", expression.GreaterThan, 1))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 3)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
			(*actually[2])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t2.Name, a[0])
		require.Equal(t, t6.Name, a[1])
		require.Equal(t, t4.Name, a[2])
	})
	t.Run("subjectID ge 1", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("subject_id", expression.GreaterThanOrEqual, 1))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 7)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
			(*actually[2])["name"].(string),
			(*actually[3])["name"].(string),
			(*actually[4])["name"].(string),
			(*actually[5])["name"].(string),
			(*actually[6])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t1.Name, a[0])
		require.Equal(t, t3.Name, a[1])
		require.Equal(t, t5.Name, a[2])
		require.Equal(t, t7.Name, a[3])
		require.Equal(t, t2.Name, a[4])
		require.Equal(t, t6.Name, a[5])
		require.Equal(t, t4.Name, a[6])
	})
	t.Run("subject.enabled eq true", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhereWithTable("Subjects", "enabled", expression.Equal, true))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t5.Name, a[0])
		require.Equal(t, t2.Name, a[1])
	})
	t.Run("lastSyncError is null", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("last_sync_error", expression.IsNull))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 1)
		require.Equal(t, m4, actually[0])
	})
	t.Run("deletedAt is not null", func(t *testing.T) {
		actually, err := ImitatorSqlWhere(items, expression.NewWhere("deleted_at", expression.IsNotNull))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 1)
		require.Equal(t, m7, actually[0])
	})
}

func TestImitatorSqlOrderBy(t *testing.T) {
	t1 := &internalTask{ID: 1, Name: "N001", IsEnabled: false, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 70), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t2 := &internalTask{ID: 2, Name: "Num2", IsEnabled: false, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 20), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: true}}}
	t3 := &internalTask{ID: 3, Name: "N003", IsEnabled: true, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 50), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t4 := &internalTask{ID: 4, Name: "num4", IsEnabled: false, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 40), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t5 := &internalTask{ID: 5, Name: "N005", IsEnabled: false, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 30), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: true}}}
	t6 := &internalTask{ID: 6, Name: "Num6", IsEnabled: true, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 60), DeletedAt: null.Time{Valid: false}, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t7 := &internalTask{ID: 7, Name: "N007", IsEnabled: false, LastSyncModifiedAt: time.Now().UTC().Add(-time.Minute * 10), DeletedAt: null.TimeFrom(time.Now().UTC()), R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}

	m1, err := RecognizeImitatorModel(t1)
	require.NoError(t, err)
	m2, err := RecognizeImitatorModel(t2)
	require.NoError(t, err)
	m3, err := RecognizeImitatorModel(t3)
	require.NoError(t, err)
	m4, err := RecognizeImitatorModel(t4)
	require.NoError(t, err)
	m5, err := RecognizeImitatorModel(t5)
	require.NoError(t, err)
	m6, err := RecognizeImitatorModel(t6)
	require.NoError(t, err)
	m7, err := RecognizeImitatorModel(t7)
	require.NoError(t, err)

	items := []*ImitatorModel{m1, m6, m2, m7, m4, m5, m3}

	t.Run("id DESC", func(t *testing.T) {
		err = ImitatorSqlOrderBy(items, expression.NewOrderBy("id", expression.Descending))
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Len(t, items, 7)
		require.Equal(t, m7, items[0])
		require.Equal(t, m6, items[1])
		require.Equal(t, m5, items[2])
		require.Equal(t, m4, items[3])
		require.Equal(t, m3, items[4])
		require.Equal(t, m2, items[5])
		require.Equal(t, m1, items[6])
	})
	t.Run("name ASC", func(t *testing.T) {
		err = ImitatorSqlOrderBy(items, expression.NewOrderBy("name", expression.Ascending))
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Len(t, items, 7)
		require.Equal(t, m1, items[0])
		require.Equal(t, m3, items[1])
		require.Equal(t, m5, items[2])
		require.Equal(t, m7, items[3])
		require.Equal(t, m2, items[4])
		require.Equal(t, m6, items[5])
		require.Equal(t, m4, items[6])
	})
	t.Run("isEnabled DESC name ASC", func(t *testing.T) {
		err = ImitatorSqlOrderBy(items, expression.NewOrderBy("is_enabled", expression.Descending), expression.NewOrderBy("id", expression.Ascending))
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Len(t, items, 7)
		require.Equal(t, m1, items[0])
		require.Equal(t, m2, items[1])
		require.Equal(t, m4, items[2])
		require.Equal(t, m5, items[3])
		require.Equal(t, m7, items[4])
		require.Equal(t, m3, items[5])
		require.Equal(t, m6, items[6])
	})
	t.Run("subject.enabled DESC name ASC", func(t *testing.T) {
		err = ImitatorSqlOrderBy(items, expression.NewOrderByWithTable("subject", "enabled", expression.Descending), expression.NewOrderBy("name", expression.Ascending))
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Len(t, items, 7)
		require.Equal(t, m1, items[0])
		require.Equal(t, m3, items[1])
		require.Equal(t, m7, items[2])
		require.Equal(t, m6, items[3])
		require.Equal(t, m4, items[4])
		require.Equal(t, m5, items[5])
		require.Equal(t, m2, items[6])
	})
	t.Run("deleted_at DESC last_sync_modified_at ASC", func(t *testing.T) {
		ImitatorSqlOrderBy(items, expression.NewOrderBy("deleted_at", expression.Ascending), expression.NewOrderBy("last_sync_modified_at", expression.Descending))
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Len(t, items, 7)
		require.Equal(t, m7, items[0])
		require.Equal(t, m2, items[1])
		require.Equal(t, m5, items[2])
		require.Equal(t, m4, items[3])
		require.Equal(t, m3, items[4])
		require.Equal(t, m6, items[5])
		require.Equal(t, m1, items[6])
	})
}

func TestImitatorSqlGroupBy(t *testing.T) {
	t1 := &internalTask{ID: 1, Name: "N001", IsEnabled: false, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t2 := &internalTask{ID: 2, Name: "Num2", IsEnabled: false, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: true}}}
	t3 := &internalTask{ID: 3, Name: "N003", IsEnabled: true, R: &internalTaskR{Subject: &internalSubject{ID: "1", Name: "Subject01", IsEnabled: false}}}
	t4 := &internalTask{ID: 4, Name: "num4", IsEnabled: false, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t5 := &internalTask{ID: 5, Name: "N005", IsEnabled: false, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: true}}}
	t6 := &internalTask{ID: 6, Name: "Num6", IsEnabled: true, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}
	t7 := &internalTask{ID: 7, Name: "N007", IsEnabled: false, R: &internalTaskR{Subject: &internalSubject{ID: "2", Name: "Subject02", IsEnabled: false}}}

	m1, err := RecognizeImitatorModel(t1)
	require.NoError(t, err)
	m2, err := RecognizeImitatorModel(t2)
	require.NoError(t, err)
	m3, err := RecognizeImitatorModel(t3)
	require.NoError(t, err)
	m4, err := RecognizeImitatorModel(t4)
	require.NoError(t, err)
	m5, err := RecognizeImitatorModel(t5)
	require.NoError(t, err)
	m6, err := RecognizeImitatorModel(t6)
	require.NoError(t, err)
	m7, err := RecognizeImitatorModel(t7)
	require.NoError(t, err)

	items := []*ImitatorModel{
		m1,
		m2,
		m3,
		m4,
		m5,
		m6,
		m7,
	}

	t.Run("is_enabled", func(t *testing.T) {
		actually, err := ImitatorSqlGroupBy(items, expression.NewGroupBy("is_enabled"))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t7.Name, a[0])
		require.Equal(t, t6.Name, a[1])
	})
	t.Run("subject.name", func(t *testing.T) {
		actually, err := ImitatorSqlGroupBy(items, expression.NewGroupWithTable("subject", "name"))
		require.NoError(t, err)
		require.NotNil(t, actually)
		require.Len(t, actually, 2)
		a := []string{
			(*actually[0])["name"].(string),
			(*actually[1])["name"].(string),
		}
		sort.Strings(a)
		require.Equal(t, t3.Name, a[0])
		require.Equal(t, t7.Name, a[1])
	})
}

func TestRecognizeImitatorModel(t *testing.T) {
	obj := internalTask{
		ID:                 rand.Int63n(0xFFFF),
		Name:               uuid.NewV4().String(),
		SubjectID:          rand.Int63n(0xFF),
		IsEnabled:          rand.Int63n(2) == 1,
		LastSyncError:      null.StringFrom(uuid.NewV4().String()),
		LastSyncModifiedAt: time.Now().UTC(),
		Metadata:           types.JSON(`{"en": "English", "ru": "Русский"}`),
		RawExternalDataset: null.JSONFrom([]byte(`{"en": "1", "ru": "2"}`)),
		DeletedAt:          null.TimeFrom(time.Now().UTC()),
	}
	m, err := RecognizeImitatorModel(obj)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, int(obj.ID), (*m)["id"])
	require.Equal(t, obj.Name, (*m)["name"])
	require.Equal(t, int(obj.SubjectID), (*m)["subject_id"])
	require.Equal(t, obj.IsEnabled, (*m)["is_enabled"])
	require.Equal(t, obj.LastSyncError.String, (*m)["last_sync_error"])
	require.Equal(t, obj.LastSyncModifiedAt, (*m)["last_sync_modified_at"])
	require.Equal(t, []byte(obj.Metadata), (*m)["metadata"])
	require.Equal(t, obj.RawExternalDataset.JSON, (*m)["raw_external_dataset"])
	require.Equal(t, obj.DeletedAt.Time, (*m)["deleted_at"])
}

func TestImitatorModelGetValue(t *testing.T) {
	t.Skip("not implemented")
}

func TestImitatorModelCompare(t *testing.T) {
	t.Skip("not implemented")
}

func TestToLowerTableName(t *testing.T) {
	require.Equal(t, "task", toLowerTableName("Task"))
	require.Equal(t, "task", toLowerTableName("Tasks"))
	require.Equal(t, "task", toLowerTableName("TASK"))
	require.Equal(t, "task", toLowerTableName("TASKS"))
	require.Equal(t, "dataset", toLowerTableName("data_set"))
}

func TestCompare(t *testing.T) {
	t.Skip("not implemented")
}

type internalTask struct {
	ID                 int64          `boil:"id"`
	Name               string         `boil:"name"`
	SubjectID          int64          `boil:"subject_id"`
	IsEnabled          bool           `boil:"is_enabled"`
	LastSyncError      null.String    `boil:"last_sync_error"`
	LastSyncModifiedAt time.Time      `boil:"last_sync_modified_at"`
	Metadata           types.JSON     `boil:"metadata"`
	RawExternalDataset null.JSON      `boil:"raw_external_dataset"`
	DeletedAt          null.Time      `boil:"deleted_at"`
	R                  *internalTaskR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L                  internalTaskL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

type internalTaskR struct {
	Subject *internalSubject `boil:"Subject"`
}

type internalSubject struct {
	ID        string `boil:"id"`
	Name      string `boil:"name"`
	IsEnabled bool   `boil:"enabled"`
}

type internalTaskL struct {
	Subject *internalSubject `boil:"Object"`
}

//func TestSimulatePSQL(t *testing.T) {
//	poll := NewPool()
//	defer func(c io.Closer) { require.NoError(t, c.Close()) }(poll)
//
//	db, err := poll.NewSQLite3()
//	require.NoError(t, err)
//	require.NotNil(t, db)
//
//	_, err = db.Exec("CREATE TABLE demo (id INTEGER PRIMARY KEY, value TEXT);")
//	require.NoError(t, err)
//
//	_, err = db.Exec("INSERT INTO demo (id, value) VALUES (1, 'test');")
//	require.NoError(t, err)
//
//	_, err = db.Exec("INSERT INTO demo (id, value) VALUES (2, '2024-11-11T01:02:03Z');")
//	require.NoError(t, err)
//
//	rows, err := db.Query("SELECT id FROM demo WHERE id = 2;")
//	require.NoError(t, err)
//	require.NotNil(t, rows)
//	defer func(c io.Closer) { require.NoError(t, c.Close()) }(rows)
//
//	mRows := &internal.RowsMock{Rows: rows}
//
//	var id int
//	var value time.Time
//	require.True(t, mRows.Next())
//	err = mRows.Scan(&id)
//	require.NoError(t, err)
//	require.Equal(t, "2024-11-11T01:02:03Z", value.Format(time.RFC3339))
//	require.Equal(t, 2, id)
//	require.False(t, mRows.Next())
//}
