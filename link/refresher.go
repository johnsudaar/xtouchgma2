package link

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

const (
	refreshMinInterval = 10 * time.Second
	refreshMaxInterval = 30 * time.Second
	inhibitInterval    = 5 * time.Second
)

type Refresher[T any] struct {
	id string

	params RefreshParams

	changedAt time.Time

	lastRefreshedAt time.Time
	currentValue    T
	nextValue       T
	inhibitedAt     time.Time

	refreshFunc func(context.Context, T) error

	lock *sync.Mutex
}

type RefreshParams struct {
	MinRefreshInterval time.Duration
	MaxRefreshInterval time.Duration
	InhibitTime        time.Duration
}

func NewRefresher[T any](id string, params RefreshParams, refreshFunc func(context.Context, T) error) *Refresher[T] {
	return &Refresher[T]{
		id:     id,
		params: params,

		refreshFunc: refreshFunc,
		lock:        &sync.Mutex{},
	}
}

func (r *Refresher[T]) Refresh(ctx context.Context) {
	r.lock.Lock()
	defer r.lock.Unlock()

	log := logger.Get(ctx).WithField("refresh_id", r.id)
	if r.changedAt.IsZero() {
		log.Debug("Initial value not ready")
		return
	}

	if time.Since(r.inhibitedAt) < r.params.InhibitTime {
		log.Debug("Refresh inhibited")
		return
	}

	// If the value was not yet sent or if it changed.
	if r.lastRefreshedAt.IsZero() || r.changedAt.After(r.lastRefreshedAt) {
		log.Debug("Value changed")
		err := r.run(ctx)
		if err != nil {
			log.WithError(err).Error("fail to refresh value when value changed")
		}
		return
	}

	randInterval := rand.Int63n(int64(r.params.MaxRefreshInterval-r.params.MinRefreshInterval)) + int64(r.params.MinRefreshInterval)
	if time.Since(r.lastRefreshedAt) < time.Duration(randInterval) {
		log.Debug("Refresh interval not reached")
		return
	}

	log.Debug("Interval reached")
	err := r.run(ctx)
	if err != nil {
		log.WithError(err).Error("fail to refresh value when interval reached")
	}
}

func (r *Refresher[T]) Inhibit() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.inhibitedAt = time.Now()
}

func (r *Refresher[T]) Set(value T) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.nextValue = value
	r.changedAt = time.Now()
}

func (r *Refresher[T]) ForceRefresh() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.lastRefreshedAt = time.Time{}
}

func (r *Refresher[T]) Value() T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.currentValue
}

func (r *Refresher[T]) run(ctx context.Context) error {
	err := r.refreshFunc(ctx, r.nextValue)
	if err != nil {
		return errors.Wrap(err, "run refresh method")
	}

	r.lastRefreshedAt = time.Now()
	r.currentValue = r.nextValue
	return nil
}
