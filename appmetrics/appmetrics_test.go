package appmetrics

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	type testCase struct {
		name             string
		livenessProbeFn  func() error
		readinessProbeFn func() error
		startupProbeFn   func() error
		url              string
		expCode          int
		expData          map[string]any
	}

	tcs := []testCase{
		// Happy path
		{
			name:            "liveness probe",
			livenessProbeFn: func() error { return nil },
			url:             livenessProbeURL,
			expCode:         http.StatusOK,
			expData: map[string]any{
				"status": "ok",
			},
		},
		{
			name:             "readiness probe",
			readinessProbeFn: func() error { return nil },
			url:              readinessProbeURL,
			expCode:          http.StatusOK,
			expData: map[string]any{
				"status": "ok",
			},
		},
		{
			name:           "startup probe",
			startupProbeFn: func() error { return nil },
			url:            startupProbeURL,
			expCode:        http.StatusOK,
			expData: map[string]any{
				"status": "ok",
			},
		},

		// Not implemented
		{
			name:             "liveness probe not implemented",
			readinessProbeFn: func() error { return nil },
			startupProbeFn:   func() error { return nil },
			url:              livenessProbeURL,
			expCode:          http.StatusNotImplemented,
			expData: map[string]any{
				"status": "failed",
				"error":  "not implemented: check function is not provided",
			},
		},
		{
			name:            "readiness probe not implemented",
			livenessProbeFn: func() error { return nil },
			startupProbeFn:  func() error { return nil },
			url:             readinessProbeURL,
			expCode:         http.StatusNotImplemented,
			expData: map[string]any{
				"status": "failed",
				"error":  "not implemented: check function is not provided",
			},
		},
		{
			name:             "startup probe not implemented",
			livenessProbeFn:  func() error { return nil },
			readinessProbeFn: func() error { return nil },
			url:              startupProbeURL,
			expCode:          http.StatusNotImplemented,
			expData: map[string]any{
				"status": "failed",
				"error":  "not implemented: check function is not provided",
			},
		},

		// Check error
		{
			name:            "liveness probe error",
			livenessProbeFn: func() error { return errors.New("blah") },
			url:             livenessProbeURL,
			expCode:         http.StatusServiceUnavailable,
			expData: map[string]any{
				"status": "failed",
				"error":  "blah",
			},
		},
		{
			name:             "readiness probe error",
			readinessProbeFn: func() error { return errors.New("blah") },
			url:              readinessProbeURL,
			expCode:          http.StatusServiceUnavailable,
			expData: map[string]any{
				"status": "failed",
				"error":  "blah",
			},
		},
		{
			name:           "startup probe error",
			startupProbeFn: func() error { return errors.New("blah") },
			url:            startupProbeURL,
			expCode:        http.StatusServiceUnavailable,
			expData: map[string]any{
				"status": "failed",
				"error":  "blah",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			e := echo.New()
			e.Use(middleware.Logger())
			e.Use(middleware.Recover())

			appMetrics := New(tc.livenessProbeFn, tc.readinessProbeFn, tc.startupProbeFn)
			appMetrics.Register(e)

			srv := httptest.NewServer(e)
			defer srv.Close()

			ctx := context.TODO()

			code, v, err := get(ctx, srv.URL+tc.url)
			r.NoError(err)
			r.Equal(tc.expCode, code)
			r.Equal(tc.expData, v)
		})
	}
}

func TestMetrics(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	appMetrics := New(nil, nil, nil)
	appMetrics.Register(e)

	srv := httptest.NewServer(e)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, srv.URL+metricsURL, nil)
	r.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	r.NoError(err)
	r.True(strings.HasPrefix(string(data), "# HELP"))
}

func get(ctx context.Context, url string) (int, map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	v := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, v, nil
}
