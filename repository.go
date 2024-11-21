package sqlinjector

import (
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/prorochestvo/sqlinjector/internal/sandbox"
	"golang.org/x/exp/constraints"
	"sync"
)

// Repository is a interface for CRUD operations of dataset
type Repository[DATAKEY constraints.Ordered, DATASET any] interface {
	Count(...Expression) (int64, error)
	ObtainAll(...Expression) (items []*DATASET, err error)
	ObtainOne(DATAKEY, ...Expression) (*DATASET, error)
	Create(*DATASET, ...*DATASET) error
	CreateOrUpdate(*DATASET, ...*DATASET) error
	Update(*DATASET, ...*DATASET) error
	Delete(*DATASET, ...*DATASET) error
	Erase(DATAKEY) error
	UpdateAll(map[string]interface{}, ...Expression) error
	DeleteAll(...Expression) error
}

// NewDummySqlBoilerRepository creates new Repository with dummy data for testing
func NewDummySqlBoilerRepository[DATAKEY constraints.Ordered, DATASET any](items ...*DATASET) (*DummyRepository[DATAKEY, DATASET], error) {
	obtainID := func(model *DATASET) (res DATAKEY, err error) {
		defaultKeyNames := []string{"id", "ID", "Id", "iD", "_id"}
		m, err := sandbox.RecognizeImitatorModel(model)
		if err != nil {
			err = fmt.Errorf("id field not recognized: %w", err)
			return
		}
		var val interface{}
		var exists bool
		for _, n := range defaultKeyNames {
			if val, exists = m.GetValue("", n); exists {
				break
			}
		}
		if !exists {
			err = fmt.Errorf("id field not recognized into %T(%v)", model, model)
			return
		}

		res, exists = val.(DATAKEY)
		if !exists {
			err = fmt.Errorf("%T is incorrect id field into %T(%v)", val, model, model)
		}

		return
	}
	obtainItems := func(items map[DATAKEY]*DATASET, expressions []Expression) (res []*DATASET, err error) {
		var where []*expression.Where
		var groupBy []*expression.GroupBy
		var orderBy []*expression.OrderBy
		for _, e := range expressions {
			if w, ok := e.(*expression.Where); ok {
				where = append(where, w)
			}
			if g, ok := e.(*expression.GroupBy); ok {
				groupBy = append(groupBy, g)
			}
			if o, ok := e.(*expression.OrderBy); ok {
				orderBy = append(orderBy, o)
			}
		}
		return sandbox.ImitatorSql(items, where, groupBy, orderBy)
	}
	dataset := make(map[DATAKEY]*DATASET)
	for i, item := range items {
		id, err := obtainID(item)
		if err != nil {
			return nil, fmt.Errorf("item[%d].id field not recognized: %w", i, err)
		}
		dataset[id] = item
	}
	return &DummyRepository[DATAKEY, DATASET]{entities: dataset, Extractor: obtainID, Filtrator: obtainItems}, nil
}

// DummyRepository is a implementation of Repository with dummy data for testing
type DummyRepository[DATAKEY constraints.Ordered, DATASET any] struct {
	m                      sync.RWMutex
	entities               map[DATAKEY]*DATASET
	Extractor              func(*DATASET) (DATAKEY, error)
	Filtrator              func(map[DATAKEY]*DATASET, []Expression) ([]*DATASET, error)
	OnBeforeCreate         func(*DATASET) error
	OnBeforeCreateOrUpdate func(*DATASET) error
	OnBeforeUpdate         func(*DATASET) error
	OnBeforeDelete         func(*DATASET) error
	OnAfterCreate          func(*DATASET) error
	OnAfterCreateOrUpdate  func(*DATASET) error
	OnAfterUpdate          func(*DATASET) error
	OnAfterDelete          func(*DATASET) error
}

// Count returns count of entities from Repository
func (r *DummyRepository[DATAKEY, DATASET]) Count(expressions ...Expression) (int64, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.entities == nil {
		return 0, nil
	}

	var where []Expression
	for _, e := range expressions {
		if _, ok := e.(*expression.Where); ok {
			where = append(where, e)
		}
	}

	items, err := r.Filtrator(r.entities, where)
	if err != nil {
		return 0, err
	}

	return int64(len(items)), nil
}

// ObtainAll returns all entities from Repository
func (r *DummyRepository[DATAKEY, DATASET]) ObtainAll(expressions ...Expression) ([]*DATASET, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.entities == nil {
		return nil, nil
	}

	items, err := r.Filtrator(r.entities, expressions)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// ObtainOne returns one item from Repository by key
func (r *DummyRepository[DATAKEY, DATASET]) ObtainOne(key DATAKEY, _ ...Expression) (*DATASET, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.entities == nil {
		return nil, fmt.Errorf("entities is empty")
	}

	item, ok := r.entities[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return item, nil
}

// Create creates new entity in Repository
func (r *DummyRepository[DATAKEY, DATASET]) Create(model *DATASET, moreModels ...*DATASET) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.entities == nil {
		r.entities = make(map[DATAKEY]*DATASET)
	}

	for i := -1; i < len(moreModels); i++ {
		if i >= 0 {
			model = moreModels[i]
		}

		if r.OnBeforeCreate != nil {
			if err := r.OnBeforeCreate(model); err != nil {
				return err
			}
		}

		id, err := r.Extractor(model)
		if err != nil {
			return err
		}
		if _, exists := r.entities[id]; exists {
			return fmt.Errorf("%v already exists", id)
		}
		r.entities[id] = model

		if r.OnAfterCreate != nil {
			if err = r.OnAfterCreate(model); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateOrUpdate creates new entity in Repository or updates existing item
func (r *DummyRepository[DATAKEY, DATASET]) CreateOrUpdate(model *DATASET, moreModels ...*DATASET) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.entities == nil {
		r.entities = make(map[DATAKEY]*DATASET)
	}

	for i := -1; i < len(moreModels); i++ {
		if i >= 0 {
			model = moreModels[i]
		}

		if r.OnBeforeCreateOrUpdate != nil {
			if err := r.OnBeforeCreateOrUpdate(model); err != nil {
				return err
			}
		}

		id, err := r.Extractor(model)
		if err != nil {
			return err
		}
		r.entities[id] = model

		if r.OnAfterCreateOrUpdate != nil {
			if err = r.OnAfterCreateOrUpdate(model); err != nil {
				return err
			}
		}
	}

	return nil
}

// Update updates existing entity in Repository
func (r *DummyRepository[DATAKEY, DATASET]) Update(model *DATASET, moreModels ...*DATASET) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.entities == nil {
		return fmt.Errorf("entities is empty")
	}

	for i := -1; i < len(moreModels); i++ {
		if i >= 0 {
			model = moreModels[i]
		}

		if r.OnBeforeUpdate != nil {
			if err := r.OnBeforeUpdate(model); err != nil {
				return err
			}
		}

		id, err := r.Extractor(model)
		if err != nil {
			return err
		}
		if _, exists := r.entities[id]; !exists {
			return fmt.Errorf("not found")
		}
		r.entities[id] = model

		if r.OnAfterUpdate != nil {
			if err := r.OnAfterUpdate(model); err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete deletes existing item in Repository
func (r *DummyRepository[DATAKEY, DATASET]) Delete(model *DATASET, moreModels ...*DATASET) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.entities == nil {
		return fmt.Errorf("entities is empty")
	}

	for i := -1; i < len(moreModels); i++ {
		if i >= 0 {
			model = moreModels[i]
		}

		if r.OnBeforeDelete != nil {
			if err := r.OnBeforeDelete(model); err != nil {
				return err
			}
		}

		id, err := r.Extractor(model)
		if err != nil {
			return err
		}
		if _, exists := r.entities[id]; !exists {
			return fmt.Errorf("not found")
		}
		r.entities[id] = nil
		delete(r.entities, id)

		if r.OnAfterDelete != nil {
			if err := r.OnAfterDelete(model); err != nil {
				return err
			}
		}
	}

	if len(r.entities) == 0 {
		r.entities = make(map[DATAKEY]*DATASET)
	}

	return nil
}

// Erase deletes existing item in Repository
func (r *DummyRepository[DATAKEY, DATASET]) Erase(key DATAKEY) error {
	item, err := r.ObtainOne(key)
	if err != nil {
		return err
	}
	return r.Delete(item)
}

// UpdateAll updates all entities in Repository
func (r *DummyRepository[DATAKEY, DATASET]) UpdateAll(m map[string]interface{}, expressions ...Expression) error {
	items, err := r.ObtainAll(expressions...)
	if err != nil {
		return err
	}
	for _, item := range items {
		err = sandbox.Merge(item, m)
		if err != nil {
			return fmt.Errorf("merge error for %v: %w", item, err)
		}
		err = r.Update(item)
		if err != nil {
			return fmt.Errorf("update error for %v: %w", item, err)
		}
	}
	return nil
}

// DeleteAll deletes all entities in Repository
func (r *DummyRepository[DATAKEY, DATASET]) DeleteAll(expressions ...Expression) error {
	items, err := r.ObtainAll(expressions...)
	if err != nil {
		return err
	}
	for _, item := range items {
		err = r.Delete(item)
		if err != nil {
			return fmt.Errorf("delete error for %v: %w", item, err)
		}
	}
	return nil
}
