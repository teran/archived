package main

import (
	"context"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	htmlPresenter "github.com/teran/archived/presenter/access/html"
	"github.com/teran/archived/service"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
	Addr            string    `envconfig:"ADDR" default:":8080"`
	LogLevel        log.Level `envconfig:"LOG_LEVEL" default:"info"`
	HTMLTemplateDir string    `envconfig:"HTML_TEMPLATE_DIR"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-access (%s @ %s) ...", appVersion, buildTimestamp)

	g, _ := errgroup.WithContext(context.Background())

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// svc := service.NewAccessService(nil, nil)
	svc := service.NewMock()
	svc.On("ListContainers").Return([]string{"test-container-1", "test-container-2"}, nil)
	svc.On("ListVersions", "test-container-1").Return([]string{"20240706101013", "202407061101314"}, nil)
	svc.On("ListObjects", "test-container-1", "20240706101013").Return([]string{
		"rockylinux/9/appstream/Packages/r/rocky-indexhtml-9.0-2.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rocky-logos-ipa-90.15-2.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rocky-logos-httpd-90.15-2.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rocky-backgrounds-90.15-2.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-srpm-macros-17-4.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/realtime-tests-2.6-5.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpmdevtools-9.5-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rpmlint-1.11-19.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redhat-text-fonts-4.0.3-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redhat-mono-fonts-4.0.3-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redhat-display-fonts-4.0.3-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-mpi-hooks-8-3.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rhel-system-roles-1.23.0-2.21.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rtkit-0.11-28.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rshim-2.0.8-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rdma-core-devel-48.0-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/redhat-rpm-config-207-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/regexp-1.5-37.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redland-1.0.17-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-plugin-systemd-inhibit-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-plugin-syslog-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-plugin-ima-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-plugin-fapolicyd-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-devel-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-cron-4.16.1.3-29.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-build-4.16.1.3-29.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-apidocs-4.16.1.3-29.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redfish-finder-0.4-9.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rig-1.1-6.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/runc-1.1.12-2.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/redis-doc-6.2.7-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/redis-devel-6.2.7-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/redis-6.2.7-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygems-devel-3.2.33-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygems-3.2.33-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-typeprof-0.15.2-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-test-unit-3.3.7-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-rss-0.2.9-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-rexml-3.2.5-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-rdoc-6.3.4.1-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-rbs-1.4.0-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-rake-13.0.3-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-power_assert-1.2.1-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-minitest-5.14.2-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-irb-1.3.5-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-bundler-2.2.33-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/ruby-default-gems-3.0.7-162.el9_4.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/radvd-2.19-5.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rasdaemon-0.6.7-9.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rrdtool-perl-1.7.2-21.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rrdtool-1.7.2-21.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/raptor2-2.0.15-30.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rear-2.6-24.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rustfmt-1.75.0-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rust-std-static-wasm32-wasi-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-std-static-wasm32-unknown-unknown-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-std-static-1.75.0-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rust-doc-1.75.0-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rust-1.75.0-1.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rust-src-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-lldb-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-gdb-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rust-debugger-common-1.75.0-1.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-udpspoof-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-snmp-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-relp-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-pgsql-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-openssl-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-omamqp1-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mysql-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmsnmptrapd-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmnormalize-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmkubernetes-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmjsonparse-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmfields-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-mmaudit-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-logrotate-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-kafka-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-gssapi-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-gnutls-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-elasticsearch-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-doc-8.2310.0-4.el9.noarch.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-crypto-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rsyslog-8.2310.0-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rest-0.8.1-11.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rasqal-0.9.33-18.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-mysql2-0.5.3-11.el9_0.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-ostree-libs-2024.3-4.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpm-ostree-2024.3-4.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/readline-devel-8.1-4.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-psych-3.3.2-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-json-2.5.1-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-io-console-0.5.7-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-bigdecimal-3.0.0-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/ruby-libs-3.0.7-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/ruby-devel-3.0.7-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/ruby-3.0.7-162.el9_4.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rpcgen-1.4-9.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rubygem-pg-1.2.3-7.el9.x86_64.rpm",
		"rockylinux/9/appstream/Packages/r/rocky-logos-90.15-2.el9.x86_64.rpm",
	}, nil)

	svc.On("GetObjectURL", "test-container-1", "20240706101013", "rockylinux/9/appstream/Packages/r/rocky-logos-90.15-2.el9.x86_64.rpm").Return("http://wikipedia.org", nil)

	p := htmlPresenter.New(svc, cfg.HTMLTemplateDir)
	p.Register(e)

	g.Go(func() error {
		srv := &http.Server{
			Addr:    cfg.Addr,
			Handler: e,
		}

		return srv.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
