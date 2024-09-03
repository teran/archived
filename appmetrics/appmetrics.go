package appmetrics

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	livenessProbeURL  = "/healthz/liveness"
	readinessProbeURL = "/healthz/readiness"
	startupProbeURL   = "/healthz/startup"
	metricsURL        = "/metrics"
)

type AppMetrics interface {
	Register(e *echo.Echo)
}

type appMetrics struct {
	livenessProbeFn  func() error
	readinessProbeFn func() error
	startupProbeFn   func() error
}

func New(livenessProbeFn, readinessProbeFn, startupProbeFn func() error) AppMetrics {
	return &appMetrics{
		livenessProbeFn:  livenessProbeFn,
		readinessProbeFn: readinessProbeFn,
		startupProbeFn:   startupProbeFn,
	}
}

func (m *appMetrics) livenessProbe(c echo.Context) error {
	if m.livenessProbeFn == nil {
		return c.JSON(http.StatusNotImplemented, echo.Map{
			"status": "failed", "error": "not implemented: check function is not provided",
		})
	}

	if err := m.livenessProbeFn(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"status": "failed", "error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func (m *appMetrics) readinessProbe(c echo.Context) error {
	if m.readinessProbeFn == nil {
		return c.JSON(http.StatusNotImplemented, echo.Map{
			"status": "failed", "error": "not implemented: check function is not provided",
		})
	}

	if err := m.readinessProbeFn(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"status": "failed", "error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func (m *appMetrics) startupProbe(c echo.Context) error {
	if m.startupProbeFn == nil {
		return c.JSON(http.StatusNotImplemented, echo.Map{
			"status": "failed", "error": "not implemented: check function is not provided",
		})
	}

	if err := m.startupProbeFn(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"status": "failed", "error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func (m *appMetrics) metrics(c echo.Context) error {
	return echo.WrapHandler(promhttp.Handler())(c)
}

func (m *appMetrics) Register(e *echo.Echo) {
	e.GET(livenessProbeURL, m.livenessProbe)
	e.GET(readinessProbeURL, m.readinessProbe)
	e.GET(startupProbeURL, m.startupProbe)
	e.GET(metricsURL, m.metrics)
}
