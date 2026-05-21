package exporter

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const DefaultSnapshotRefreshInterval = 5 * time.Minute

type Snapshotter[T any] interface {
	Snapshot(context.Context, time.Time) T
}

type SnapshotStatus struct {
	AttemptTime time.Time
	Success     bool
}

type SnapshotCollectorOptions[T any] struct {
	Namespace       string
	Logger          *slog.Logger
	Snapshotter     Snapshotter[T]
	RefreshInterval time.Duration
	StatusFunc      func(T) SnapshotStatus
	DescribeFunc    func(chan<- *prometheus.Desc)
	CollectFunc     func(chan<- prometheus.Metric, T, time.Time)
	ErrorLogFunc    func(*slog.Logger, T)
	Now             func() time.Time

	LastCollectionSuccessHelp    string
	LastCollectionTimestampHelp  string
	LastSuccessfulCollectionHelp string
}

type SnapshotCollector[T any] struct {
	namespace       string
	logger          *slog.Logger
	snapshotter     Snapshotter[T]
	refreshInterval time.Duration
	statusFunc      func(T) SnapshotStatus
	describeFunc    func(chan<- *prometheus.Desc)
	collectFunc     func(chan<- prometheus.Metric, T, time.Time)
	errorLogFunc    func(*slog.Logger, T)
	now             func() time.Time

	mu                       sync.Mutex
	cond                     *sync.Cond
	initialized              bool
	backgroundStarted        bool
	refreshing               bool
	snapshot                 T
	snapshotStatus           SnapshotStatus
	lastSuccessfulCollection time.Time

	lastCollectionSuccessDesc    *prometheus.Desc
	lastCollectionTimestampDesc  *prometheus.Desc
	lastSuccessfulCollectionDesc *prometheus.Desc
}

func NewSnapshotCollector[T any](options SnapshotCollectorOptions[T]) *SnapshotCollector[T] {
	namespace := options.Namespace
	if namespace == "" {
		namespace = "exporter"
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	snapshotter := options.Snapshotter
	if snapshotter == nil {
		snapshotter = zeroSnapshotter[T]{}
	}
	refreshInterval := options.RefreshInterval
	if refreshInterval <= 0 {
		refreshInterval = DefaultSnapshotRefreshInterval
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	statusFunc := options.StatusFunc
	if statusFunc == nil {
		statusFunc = zeroSnapshotStatus[T]
	}

	collector := &SnapshotCollector[T]{
		namespace:       namespace,
		logger:          logger,
		snapshotter:     snapshotter,
		refreshInterval: refreshInterval,
		statusFunc:      statusFunc,
		describeFunc:    options.DescribeFunc,
		collectFunc:     options.CollectFunc,
		errorLogFunc:    options.ErrorLogFunc,
		now:             now,

		lastCollectionSuccessDesc: prometheus.NewDesc(
			namespace+"_last_collection_success",
			defaultString(options.LastCollectionSuccessHelp, "Whether the last collection succeeded"),
			nil,
			nil,
		),
		lastCollectionTimestampDesc: prometheus.NewDesc(
			namespace+"_last_collection_timestamp_seconds",
			defaultString(options.LastCollectionTimestampHelp, "Unix timestamp of the last collection attempt"),
			nil,
			nil,
		),
		lastSuccessfulCollectionDesc: prometheus.NewDesc(
			namespace+"_last_successful_collection_timestamp_seconds",
			defaultString(options.LastSuccessfulCollectionHelp, "Unix timestamp of the last successful collection"),
			nil,
			nil,
		),
	}
	collector.cond = sync.NewCond(&collector.mu)
	return collector
}

func (c *SnapshotCollector[T]) Describe(ch chan<- *prometheus.Desc) {
	if c.describeFunc != nil {
		c.describeFunc(ch)
	}
	ch <- c.lastCollectionSuccessDesc
	ch <- c.lastCollectionTimestampDesc
	ch <- c.lastSuccessfulCollectionDesc
}

func (c *SnapshotCollector[T]) Start(ctx context.Context) {
	c.mu.Lock()
	if c.backgroundStarted {
		c.mu.Unlock()
		return
	}
	c.backgroundStarted = true
	c.mu.Unlock()

	go c.refreshLoop(ctx)
}

func (c *SnapshotCollector[T]) Collect(ch chan<- prometheus.Metric) {
	now := c.now()
	state := c.currentSnapshot(now)

	if c.collectFunc != nil {
		c.collectFunc(ch, state.snapshot, now)
	}

	ch <- prometheus.MustNewConstMetric(
		c.lastCollectionSuccessDesc,
		prometheus.GaugeValue,
		BoolFloat(state.status.Success),
	)
	ch <- prometheus.MustNewConstMetric(
		c.lastCollectionTimestampDesc,
		prometheus.GaugeValue,
		UnixTimestamp(state.status.AttemptTime),
	)
	ch <- prometheus.MustNewConstMetric(
		c.lastSuccessfulCollectionDesc,
		prometheus.GaugeValue,
		UnixTimestamp(state.lastSuccessfulCollection),
	)
}

func (c *SnapshotCollector[T]) refreshLoop(ctx context.Context) {
	c.refresh(ctx, c.now())

	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refresh(ctx, c.now())
		}
	}
}

func (c *SnapshotCollector[T]) refresh(ctx context.Context, now time.Time) {
	if !c.beginRefresh() {
		return
	}

	snapshot := c.snapshotter.Snapshot(ctx, now)
	c.logSnapshotErrors(snapshot)
	c.finishRefresh(snapshot)
}

func (c *SnapshotCollector[T]) currentSnapshot(now time.Time) snapshotState[T] {
	for {
		c.mu.Lock()
		if c.initialized && (c.backgroundStarted || now.Sub(c.snapshotStatus.AttemptTime) < c.refreshInterval) {
			state := c.snapshotStateLocked()
			c.mu.Unlock()
			return state
		}
		if c.refreshing {
			for c.refreshing {
				c.cond.Wait()
			}
			c.mu.Unlock()
			continue
		}
		c.refreshing = true
		c.mu.Unlock()

		snapshot := c.snapshotter.Snapshot(context.Background(), now)
		c.logSnapshotErrors(snapshot)

		c.mu.Lock()
		c.storeSnapshotLocked(snapshot)
		c.refreshing = false
		c.cond.Broadcast()
		state := c.snapshotStateLocked()
		c.mu.Unlock()
		return state
	}
}

func (c *SnapshotCollector[T]) beginRefresh() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.refreshing {
		return false
	}
	c.refreshing = true
	return true
}

func (c *SnapshotCollector[T]) finishRefresh(snapshot T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storeSnapshotLocked(snapshot)
	c.refreshing = false
	c.cond.Broadcast()
}

func (c *SnapshotCollector[T]) storeSnapshotLocked(snapshot T) {
	status := c.statusFunc(snapshot)
	if status.Success {
		c.lastSuccessfulCollection = status.AttemptTime
	}

	c.snapshot = snapshot
	c.snapshotStatus = status
	c.initialized = true
}

func (c *SnapshotCollector[T]) snapshotStateLocked() snapshotState[T] {
	return snapshotState[T]{
		snapshot:                 c.snapshot,
		status:                   c.snapshotStatus,
		lastSuccessfulCollection: c.lastSuccessfulCollection,
	}
}

func (c *SnapshotCollector[T]) logSnapshotErrors(snapshot T) {
	if c.errorLogFunc != nil {
		c.errorLogFunc(c.logger, snapshot)
	}
}

type snapshotState[T any] struct {
	snapshot                 T
	status                   SnapshotStatus
	lastSuccessfulCollection time.Time
}

type zeroSnapshotter[T any] struct{}

func (zeroSnapshotter[T]) Snapshot(context.Context, time.Time) T {
	var snapshot T
	return snapshot
}

func zeroSnapshotStatus[T any](T) SnapshotStatus {
	return SnapshotStatus{}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
