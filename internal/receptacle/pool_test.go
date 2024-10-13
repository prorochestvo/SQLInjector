package receptacle

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewContainerPoll(t *testing.T) {
	pool := NewContainerPoll()
	require.NotNil(t, pool)
	require.NotNil(t, pool.containers)
	require.Len(t, pool.containers, 0)
}

func TestContainerPoll_Append(t *testing.T) {
	pool := NewContainerPoll()
	container := &TestContainer{}

	err := pool.Append("testContainer", container, "D1", "S1")
	require.NoError(t, err)

	err = pool.Append("testContainer", container, "D2", "S2")
	require.Error(t, err)
	require.Contains(t, err.Error(), "container already exists")
}

func TestContainerPoll_Get(t *testing.T) {
	pool := NewContainerPoll()
	container := &TestContainer{}

	_, _, err := pool.Get("nonExistentContainer")
	require.Error(t, err)
	require.Equal(t, ErrContainerNotFound, err)

	_ = pool.Append("testContainer", container, "D3", "S3")
	c, a, err := pool.Get("testContainer")
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NotNil(t, a)
	require.Equal(t, []string{"D3", "S3"}, a)
	require.Equal(t, container, c)
}

func TestContainerPoll_Remove(t *testing.T) {
	pool := NewContainerPoll()
	container := &TestContainer{}

	err := pool.Remove("nonExistentContainer")
	require.Error(t, err)
	require.Equal(t, ErrContainerNotFound, err)

	_ = pool.Append("testContainer", container, "D4", "S4")
	err = pool.Remove("testContainer")
	require.NoError(t, err)

	_, _, err = pool.Get("testContainer")
	require.Error(t, err)
	require.Equal(t, ErrContainerNotFound, err)
}
