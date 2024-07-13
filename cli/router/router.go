package router

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrNoSuchRoute = errors.New("no such route")

type Router interface {
	Register(command string, fn func(ctx context.Context) error)
	Call(command string) error
}

type router struct {
	ctx    context.Context
	mutex  *sync.RWMutex
	routes map[string]func(ctx context.Context) error
}

func New(ctx context.Context) Router {
	return &router{
		ctx:    ctx,
		mutex:  &sync.RWMutex{},
		routes: make(map[string]func(ctx context.Context) error),
	}
}

func (r *router) Register(command string, fn func(ctx context.Context) error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log.Debugf("route `%s` registered", command)

	r.routes[command] = fn
}

func (r *router) Call(command string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	log.Tracef("route request: `%s`", command)

	if fn, ok := r.routes[command]; ok {
		return fn(r.ctx)
	}
	return ErrNoSuchRoute
}
