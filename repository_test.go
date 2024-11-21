package sqlinjector

import (
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ Repository[int, any] = &DummyRepository[int, any]{}
var _ Repository[uint, any] = &DummyRepository[uint, any]{}
var _ Repository[string, any] = &DummyRepository[string, any]{}

func TestNewRepositoryStub(t *testing.T) {
	var repo Repository[string, internalSubject]
	var err error
	repo, err = NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
}

func TestDummyRepository_Count(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	val, err := repo.Count()
	require.NoError(t, err)
	require.Equal(t, int64(7), val)
}

func TestDummyRepository_ObtainAll(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 7", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: false},
		&internalSubject{ID: "5", Name: "SubjectName 3", IsEnabled: false},
		&internalSubject{ID: "6", Name: "SubjectName 2", IsEnabled: false},
		&internalSubject{ID: "7", Name: "SubjectName 1", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("GetAll", func(t *testing.T) {
		var val []*internalSubject
		val, err = repo.ObtainAll()
		require.NoError(t, err)
		require.NotNil(t, val)
		require.Len(t, val, 7)
	})
	t.Run("Filtered", func(t *testing.T) {
		var val []*internalSubject
		val, err = repo.ObtainAll(
			expression.NewWhere("id", expression.In, "2", "3", "4", "5"),
			expression.NewGroupBy("enabled"),
			expression.NewOrderBy("name", expression.Ascending),
		)
		require.NoError(t, err)
		require.NotNil(t, val)
		require.Len(t, val, 2)
		require.Contains(t, "45", val[0].ID)
		require.Contains(t, "23", val[1].ID)
	})
}

func TestDummyRepository_ObtainOne(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("GetOne", func(t *testing.T) {
		val, err := repo.ObtainOne("6")
		require.NoError(t, err)
		require.NotNil(t, val)
		require.Contains(t, "6", val.ID)
	})
}

func TestDummyRepository_Create(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("SingleCreate", func(t *testing.T) {
		m8 := &internalSubject{ID: "8"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m8)
		require.NoError(t, err)
		require.Len(t, repo.entities, l+1)
		require.Equal(t, m8, repo.entities["8"])
	})
	t.Run("MultipleCreate", func(t *testing.T) {
		m10 := &internalSubject{ID: "10"}
		m11 := &internalSubject{ID: "11"}
		m12 := &internalSubject{ID: "12"}
		m13 := &internalSubject{ID: "13"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m10, m11, m12, m13)
		require.NoError(t, err)
		require.Len(t, repo.entities, l+4)
		require.Equal(t, m10, repo.entities["10"])
		require.Equal(t, m11, repo.entities["11"])
		require.Equal(t, m12, repo.entities["12"])
		require.Equal(t, m13, repo.entities["13"])
	})
}

func TestDummyRepository_CreateOrUpdate(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("SingleCreate", func(t *testing.T) {
		m8 := &internalSubject{ID: "8"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m8)
		require.NoError(t, err)
		require.Len(t, repo.entities, l+1)
		require.Equal(t, m8, repo.entities["8"])
	})
	t.Run("MultipleCreate", func(t *testing.T) {
		m10 := &internalSubject{ID: "10"}
		m11 := &internalSubject{ID: "11"}
		m12 := &internalSubject{ID: "12"}
		m13 := &internalSubject{ID: "13"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m10, m11, m12, m13)
		require.NoError(t, err)
		require.Len(t, repo.entities, l+4)
		require.Equal(t, m10, repo.entities["10"])
		require.Equal(t, m11, repo.entities["11"])
		require.Equal(t, m12, repo.entities["12"])
		require.Equal(t, m13, repo.entities["13"])
	})
	t.Run("SingleUpdate", func(t *testing.T) {
		m8 := &internalSubject{ID: "1"}
		err = repo.CreateOrUpdate(m8)
		l := len(repo.entities)
		require.NoError(t, err)
		require.Len(t, repo.entities, l)
		require.Equal(t, m8, repo.entities["1"])
	})
	t.Run("MultipleUpdate", func(t *testing.T) {
		m3 := &internalSubject{ID: "3"}
		m4 := &internalSubject{ID: "4"}
		m5 := &internalSubject{ID: "5"}
		m7 := &internalSubject{ID: "7"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m3, m4, m5, m7)
		require.NoError(t, err)
		require.Len(t, repo.entities, l)
		require.Equal(t, m3, repo.entities["3"])
		require.Equal(t, m4, repo.entities["4"])
		require.Equal(t, m5, repo.entities["5"])
		require.Equal(t, m7, repo.entities["7"])
	})
	t.Run("MultipleCreateAndUpdate", func(t *testing.T) {
		m2 := &internalSubject{ID: "2"}
		m3 := &internalSubject{ID: "3"}
		m4 := &internalSubject{ID: "4"}
		m7 := &internalSubject{ID: "7"}
		m20 := &internalSubject{ID: "20"}
		m21 := &internalSubject{ID: "21"}
		m22 := &internalSubject{ID: "22"}
		m23 := &internalSubject{ID: "23"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m3, m4, m22, m20, m21, m2, m23, m7)
		require.NoError(t, err)
		require.Len(t, repo.entities, l+4)
		require.Equal(t, m2, repo.entities["2"])
		require.Equal(t, m3, repo.entities["3"])
		require.Equal(t, m4, repo.entities["4"])
		require.Equal(t, m7, repo.entities["7"])
		require.Equal(t, m20, repo.entities["20"])
		require.Equal(t, m21, repo.entities["21"])
		require.Equal(t, m22, repo.entities["22"])
		require.Equal(t, m23, repo.entities["23"])
	})
}

func TestDummyRepository_Update(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("SingleUpdate", func(t *testing.T) {
		m8 := &internalSubject{ID: "1"}
		err = repo.CreateOrUpdate(m8)
		l := len(repo.entities)
		require.NoError(t, err)
		require.Len(t, repo.entities, l)
		require.Equal(t, m8, repo.entities["1"])
	})
	t.Run("MultipleUpdate", func(t *testing.T) {
		m3 := &internalSubject{ID: "3"}
		m4 := &internalSubject{ID: "4"}
		m5 := &internalSubject{ID: "5"}
		m7 := &internalSubject{ID: "7"}
		l := len(repo.entities)
		err = repo.CreateOrUpdate(m3, m4, m5, m7)
		require.NoError(t, err)
		require.Len(t, repo.entities, l)
		require.Equal(t, m3, repo.entities["3"])
		require.Equal(t, m4, repo.entities["4"])
		require.Equal(t, m5, repo.entities["5"])
		require.Equal(t, m7, repo.entities["7"])
	})
}

func TestDummyRepository_Delete(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: true},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: true},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: true},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	t.Run("SingleDelete", func(t *testing.T) {
		m2 := &internalSubject{ID: "2"}
		l := len(repo.entities)
		err = repo.Delete(m2)
		require.NoError(t, err)
		require.Len(t, repo.entities, l-1)
		_, ok := repo.entities["2"]
		require.False(t, ok)
	})
	t.Run("MultipleDelete", func(t *testing.T) {
		m3 := &internalSubject{ID: "3"}
		m4 := &internalSubject{ID: "4"}
		m5 := &internalSubject{ID: "5"}
		m7 := &internalSubject{ID: "7"}
		l := len(repo.entities)
		err = repo.Delete(m3, m4, m5, m7)
		require.NoError(t, err)
		require.Len(t, repo.entities, l-4)
		_, ok := repo.entities["3"]
		require.False(t, ok)
		_, ok = repo.entities["4"]
		require.False(t, ok)
		_, ok = repo.entities["5"]
		require.False(t, ok)
		_, ok = repo.entities["7"]
		require.False(t, ok)
	})

}

func TestDummyRepository_UpdateAll(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: false},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: false},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: false},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	err = repo.UpdateAll(
		map[string]interface{}{
			"name":    "DEMO",
			"enabled": true,
		},
		expression.NewWhere("enabled", expression.Equal, false),
	)
	require.NoError(t, err)
	require.Len(t, repo.entities, 7)
	for id, entity := range repo.entities {
		if id == "2" || id == "4" || id == "6" {
			require.Equal(t, "DEMO", entity.Name)
		} else {
			require.NotEqual(t, "DEMO", entity.Name)
		}
		require.Equal(t, true, entity.IsEnabled)
	}
}

func TestDummyRepository_DeleteAll(t *testing.T) {
	repo, err := NewDummySqlBoilerRepository[string, internalSubject](
		&internalSubject{ID: "1", Name: "SubjectName 1", IsEnabled: true},
		&internalSubject{ID: "2", Name: "SubjectName 2", IsEnabled: false},
		&internalSubject{ID: "3", Name: "SubjectName 3", IsEnabled: true},
		&internalSubject{ID: "4", Name: "SubjectName 4", IsEnabled: false},
		&internalSubject{ID: "5", Name: "SubjectName 5", IsEnabled: true},
		&internalSubject{ID: "6", Name: "SubjectName 6", IsEnabled: false},
		&internalSubject{ID: "7", Name: "SubjectName 7", IsEnabled: true},
	)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Len(t, repo.entities, 7)

	err = repo.DeleteAll(expression.NewWhere("enabled", expression.Equal, true))
	require.NoError(t, err)
	require.Len(t, repo.entities, 3)
	require.Equal(t, repo.entities["2"].Name, "SubjectName 2")
	require.Equal(t, repo.entities["4"].Name, "SubjectName 4")
	require.Equal(t, repo.entities["6"].Name, "SubjectName 6")
}

type internalSubject struct {
	ID        string `boil:"id"`
	Name      string `boil:"name"`
	IsEnabled bool   `boil:"enabled"`
}
