package receptacle

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sync"
)

func NewContainerPoll() *ContainerPoll {
	poll := &ContainerPoll{
		containers: make(map[string]*TestContainer, 2),
		sources:    make(map[string][]string, 2),
	}
	return poll
}

type ContainerPoll struct {
	containers map[string]*TestContainer
	sources    map[string][]string
	m          sync.RWMutex
}

func (cp *ContainerPoll) NewDB(name string) (*sql.DB, error) {
	cp.m.Lock()
	defer cp.m.Unlock()

	_, ok := cp.containers[name]
	if ok {
		return nil, ErrContainerNotFound
	}

	args, ok := cp.sources[name]
	if ok || len(args) != 2 {
		return nil, fmt.Errorf("invalid source for container %s", name)
	}

	db, err := sql.Open(args[0], args[1])
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (cp *ContainerPoll) Append(name string, container *TestContainer, args ...string) error {
	cp.m.Lock()
	defer cp.m.Unlock()

	_, ok := cp.containers[name]
	if ok {
		return errors.New("container already exists")
	}

	cp.containers[name] = container
	cp.sources[name] = args

	return nil
}

func (cp *ContainerPoll) Get(name string) (*TestContainer, []string, error) {
	cp.m.RLock()
	defer cp.m.RUnlock()

	c, ok := cp.containers[name]
	if !ok {
		return nil, nil, ErrContainerNotFound
	}

	s, ok := cp.sources[name]
	if !ok {
		return nil, nil, ErrContainerNotFound
	}

	return c, s, nil
}

func (cp *ContainerPoll) Remove(name string) (err error) {
	cp.m.Lock()
	defer cp.m.Unlock()

	c, ok := cp.containers[name]
	if !ok {
		err = ErrContainerNotFound
		return err
	}
	defer func(c io.Closer) { err = errors.Join(err, c.Close()) }(c)

	cp.containers[name] = nil // garbage collector optimization
	cp.sources[name] = nil    // garbage collector optimization

	delete(cp.containers, name)
	delete(cp.sources, name)

	return
}

var ErrContainerNotFound = errors.New("container not found")
