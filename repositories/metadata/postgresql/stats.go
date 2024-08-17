package postgresql

import (
	"context"

	"github.com/teran/archived/exporter/models"
)

func (r *repository) CountStats(ctx context.Context) (*models.Stats, error) {
	stats := models.Stats{}
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("COUNT(*)").
		From("containers"),
	)
	if err != nil {
		return nil, err
	}

	if err := row.Scan(&stats.ContainersCount); err != nil {
		return nil, err
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select(
			"COUNT(*) AS versions_count",
			"c.name AS container_name",
			"v.is_published AS is_published",
		).
		From("versions v").
		Join("containers c ON c.id = v.container_id").
		GroupBy("c.name", "v.name", "v.is_published").
		OrderBy("c.name", "v.name", "v.is_published"),
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		vc := models.VersionsCount{}
		if err := rows.Scan(&vc.VersionsCount, &vc.ContainerName, &vc.IsPublished); err != nil {
			return nil, err
		}

		stats.VersionsCount = append(stats.VersionsCount, vc)
	}

	rows, err = selectQuery(ctx, r.db, psql.
		Select(
			"COUNT(*)",
			"c.name AS container_name",
			"v.name AS version_name",
			"v.is_published AS is_published",
		).
		From("versions v").
		Join("containers c ON c.id = v.container_id").
		Join("objects o ON o.version_id = v.id").
		GroupBy("c.name", "v.name", "v.is_published").
		OrderBy("c.name", "v.name", "v.is_published"),
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		oc := models.ObjectsCount{}
		if err := rows.Scan(&oc.ObjectsCount, &oc.ContainerName, &oc.VersionName, &oc.IsPublished); err != nil {
			return nil, err
		}

		stats.ObjectsCount = append(stats.ObjectsCount, oc)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("COUNT(*)").
		From("blobs"),
	)
	if err != nil {
		return nil, err
	}

	if err := row.Scan(&stats.BlobsCount); err != nil {
		return nil, err
	}

	rows, err = selectQuery(ctx, r.db, psql.
		Select(
			"SUM(b.size) AS size_bytes",
			"c.name AS container_name",
			"v.name AS version_name",
			"v.is_published as is_published",
		).
		From("blobs b").
		Join("objects o ON o.blob_id = b.id").
		Join("versions v ON v.id = o.version_id").
		Join("containers c ON c.id = v.container_id").
		GroupBy("c.name", "v.name", "v.is_published").
		OrderBy("c.name", "v.name", "v.is_published"),
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		brsb := models.BlobsRawSizeBytes{}
		if err := rows.Scan(&brsb.SizeBytes, &brsb.ContainerName, &brsb.VersionName, &brsb.IsPublished); err != nil {
			return nil, err
		}

		stats.BlobsRawSizeBytes = append(stats.BlobsRawSizeBytes, brsb)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("SUM(size)").
		From("blobs"),
	)
	if err != nil {
		return nil, err
	}

	if err := row.Scan(&stats.BlobsTotalSizeBytes); err != nil {
		return nil, err
	}

	return &stats, nil
}
