package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/repositories/metadata"
)

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	namespacesTotal   *prometheus.GaugeVec
	containersTotal   *prometheus.GaugeVec
	versionsTotal     *prometheus.GaugeVec
	objectsTotal      *prometheus.GaugeVec
	blobsTotal        *prometheus.GaugeVec
	blobsSize         *prometheus.GaugeVec
	blobsTotalRawSize *prometheus.GaugeVec

	repo            metadata.Repository
	observeInterval time.Duration
	mutex           *sync.Mutex
}

func New(repo metadata.Repository, observeInterval time.Duration) (Service, error) {
	svc := &service{
		repo:            repo,
		observeInterval: observeInterval,

		namespacesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "namespaces_amount",
				Help:      "Total amount of namespaces",
			},
			[]string{},
		),

		containersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "containers_amount",
				Help:      "Total amount of containers",
			},
			[]string{},
		),

		versionsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "versions_amount",
				Help:      "Total amount of versions",
			},
			[]string{"container_namespace", "container_name", "is_published"},
		),

		objectsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "objects_amount",
				Help:      "Total amount of objects",
			},
			[]string{"container_namespace", "container_name", "version_id", "is_published"},
		),

		blobsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_amount",
				Help:      "Total amount of blobs",
			},
			[]string{},
		),

		blobsSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_raw_size_bytes",
				Help:      "Total raw size of blobs (i.e. before deduplication)",
			}, []string{"container_namespace", "container_name", "version_id", "is_published"},
		),

		blobsTotalRawSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_effective_size_total_bytes",
				Help:      "Total effective size of blobs (i.e. after deduplication)",
			}, []string{},
		),

		mutex: &sync.Mutex{},
	}

	for _, m := range []*prometheus.GaugeVec{
		svc.namespacesTotal,
		svc.containersTotal,
		svc.versionsTotal,
		svc.objectsTotal,
		svc.blobsTotal,
		svc.blobsSize,
		svc.blobsTotalRawSize,
	} {
		if err := prometheus.Register(m); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func (s *service) observe(ctx context.Context) error {
	log.Trace("running observe() to gather metrics ...")

	stats, err := s.repo.CountStats(ctx)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"namespaces": stats.NamespacesCount,
		"containers": stats.ContainersCount,
	}).Trace("publishing metrics ...")

	s.namespacesTotal.WithLabelValues().Set(float64(stats.NamespacesCount))
	s.containersTotal.WithLabelValues().Set(float64(stats.ContainersCount))

	for _, vt := range stats.VersionsCount {
		s.versionsTotal.WithLabelValues(
			vt.Namespace, vt.ContainerName, strconv.FormatBool(vt.IsPublished),
		).Set(float64(vt.VersionsCount))
	}

	for _, ot := range stats.ObjectsCount {
		s.objectsTotal.WithLabelValues(
			ot.Namespace, ot.ContainerName, ot.VersionName, strconv.FormatBool(ot.IsPublished),
		).Set(float64(ot.ObjectsCount))
	}

	s.blobsTotal.WithLabelValues().Set(float64(stats.BlobsCount))

	for _, bs := range stats.BlobsRawSizeBytes {
		s.blobsSize.WithLabelValues(
			bs.Namespace, bs.ContainerName, bs.VersionName, strconv.FormatBool(bs.IsPublished),
		).Set(float64(bs.SizeBytes))
	}

	s.blobsTotalRawSize.WithLabelValues().Set(float64(stats.BlobsTotalSizeBytes))

	return nil
}

func (s *service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.observeInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		case <-ticker.C:
			go func() {
				if !s.mutex.TryLock() {
					log.Warn("lock is already taken. Skipping the run ...")
					return
				}
				defer s.mutex.Unlock()

				if err := s.observe(ctx); err != nil {
					log.Warnf("error running observe(): %s", err)
				}
			}()
		}
	}
}
