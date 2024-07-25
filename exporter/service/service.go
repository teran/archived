package service

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teran/archived/repositories/metadata"
)

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	containersTotal   *prometheus.GaugeVec
	versionsTotal     *prometheus.GaugeVec
	objectsTotal      *prometheus.GaugeVec
	blobsTotal        *prometheus.GaugeVec
	blobsSize         *prometheus.GaugeVec
	blobsTotalRawSize *prometheus.GaugeVec

	repo metadata.Repository
}

func New(repo metadata.Repository) (Service, error) {
	svc := &service{
		repo: repo,

		containersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "containers_total",
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
			[]string{"container", "is_published"},
		),

		objectsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "objects_amount",
				Help:      "Total amount of objects",
			},
			[]string{"container", "version", "is_published"},
		),

		blobsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_total",
				Help:      "Total amount of blobs",
			},
			[]string{},
		),

		blobsSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_raw_size_bytes",
				Help:      "Total raw size of blobs (i.e. before deduplication)",
			}, []string{"container", "version", "is_published"},
		),

		blobsTotalRawSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "archived",
				Name:      "blobs_effective_size_total_bytes",
				Help:      "Total effective size of blobs (i.e. after deduplication)",
			}, []string{},
		),
	}

	for _, m := range []*prometheus.GaugeVec{
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
	stats, err := s.repo.CountStats(ctx)
	if err != nil {
		return err
	}

	s.containersTotal.WithLabelValues().Set(float64(stats.ContainersCount))
	for _, vt := range stats.VersionsCount {
		s.versionsTotal.WithLabelValues(
			vt.ContainerName, strconv.FormatBool(vt.IsPublished),
		).Set(float64(vt.VersionsCount))
	}

	for _, ot := range stats.ObjectsCount {
		s.objectsTotal.WithLabelValues(
			ot.ContainerName, ot.VersionName, strconv.FormatBool(ot.IsPublished),
		).Set(float64(ot.ObjectsCount))
	}

	s.blobsTotal.WithLabelValues().Set(float64(stats.BlobsCount))

	for _, bs := range stats.BlobsRawSizeBytes {
		s.blobsSize.WithLabelValues(
			bs.ContainerName, bs.VersionName, strconv.FormatBool(bs.IsPublished),
		).Set(float64(bs.SizeBytes))
	}

	s.blobsTotalRawSize.WithLabelValues().Set(float64(stats.BlobsTotalSizeBytes))

	return nil
}

func (s *service) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
			if err := s.observe(ctx); err != nil {
				return err
			}
		}
	}
}
