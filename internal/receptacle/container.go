package receptacle

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sync"
	"time"
)

func NewTestContainers(image string, ports []string, env map[string]string, waitPorts ...string) (tc *TestContainer, err error) {
	defer func() {
		if err != nil && tc != nil && tc.logs != nil {
			println(tc.logs.Sting())
		}
	}()

	tc = &TestContainer{logs: &internalLogging{}}

	var strategies []wait.Strategy
	for _, p := range waitPorts {
		strategies = append(strategies, wait.ForListeningPort(nat.Port(p)).WithStartupTimeout(30*time.Second))
	}
	if strategies == nil {
		strategies = append(strategies, wait.ForHealthCheck())
	}

	request := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: ports,
		Env:          env,
		WaitingFor:   wait.ForAll(strategies...),
	}

	c, err := run(request, tc.logs)
	if err != nil || c == nil {
		if err == nil {
			err = fmt.Errorf("container is nil")
		}
		return nil, err
	}

	tc.container = c

	return tc, nil
}

type TestContainer struct {
	container testcontainers.Container
	logs      *internalLogging
	m         sync.RWMutex
}

func (tc *TestContainer) MappedHostPort(port string) (string, string, error) {
	tc.m.RLock()
	defer tc.m.RUnlock()

	if tc.container == nil {
		return "", "", fmt.Errorf("container is closed")
	}

	ctx := context.Background()

	host, err := tc.container.Host(ctx)
	if err != nil {
		return "", "", err
	}

	p, err := tc.container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return "", "", err
	}

	return host, p.Port(), nil
}

func (tc *TestContainer) Close() error {
	tc.m.Lock()
	defer tc.m.Unlock()

	if tc.container == nil {
		return fmt.Errorf("container is closed")
	}

	err := tc.container.Terminate(context.Background())
	if err != nil {
		return err
	}

	tc.container = nil

	return nil
}

func run(request testcontainers.ContainerRequest, logging testcontainers.Logging) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
		Logger:           logging,
	})
	if err != nil || container == nil {
		if err == nil {
			err = fmt.Errorf("container is nil")
		}
		return nil, err
	}
	return container, nil
}

type internalLogging struct {
	buffer bytes.Buffer
	m      sync.RWMutex
}

func (l *internalLogging) Sting() string {
	l.m.RLock()
	defer l.m.RUnlock()

	return l.buffer.String()
}

func (l *internalLogging) Printf(format string, v ...interface{}) {
	l.m.Lock()
	defer l.m.Unlock()

	_, _ = fmt.Fprintf(&l.buffer, format, v...)
}
