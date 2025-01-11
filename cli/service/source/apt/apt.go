package apt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/lazyblob"
	"github.com/teran/archived/cli/service/source"
)

var (
	_ source.Source = (*repository)(nil)

	ErrChecksumMismatch = errors.New("checksum mismatch")
)

type repository struct {
	repoURL       string
	suites        []string
	components    []string
	architectures []string
}

func New(repoURL string, suites, components, architectures []string) source.Source {
	log.WithFields(log.Fields{
		"url": repoURL,
	}).Trace("initializing YUM source ...")

	return &repository{
		repoURL:       repoURL,
		suites:        suites,
		components:    components,
		architectures: architectures,
	}
}

func (r *repository) Process(ctx context.Context, handler source.ObjectHandler) error {
	log.WithFields(log.Fields{
		"repository_url": r.repoURL,
	}).Info("running creating version from APT repository ...")

	type metadataBlob struct {
		path     string
		contents []byte
		sha256   string
		size     uint64
	}

	for _, suite := range r.suites {
		log.WithFields(log.Fields{
			"suite": suite,
		}).Info("processing suite ...")

		log.WithFields(log.Fields{
			"suite": suite,
		}).Debug("processing Release file ...")

		for _, filename := range []string{
			fmt.Sprintf("dists/%s/ChangeLog", suite),
			fmt.Sprintf("dists/%s/InRelease", suite),
			fmt.Sprintf("dists/%s/Release", suite),
			fmt.Sprintf("dists/%s/Release.gpg", suite),
		} {
			data, err := getFile(ctx, fmt.Sprintf("%s/%s", r.repoURL, filename))
			if err != nil {
				return err
			}

			checksum, err := sha256FromBytes(data)
			if err != nil {
				return err
			}

			if err := handler(ctx, source.Object{
				Path: filename,
				Contents: func(ctx context.Context) (io.Reader, error) {
					return bytes.NewReader(data), nil
				},
				SHA256:   checksum,
				Size:     uint64(len(data)),
				MimeType: http.DetectContentType(data),
			}); err != nil {
				return err
			}
		}

		for _, component := range r.components {
			for _, architecture := range r.architectures {
				if err := func(component, architecture string) error {
					log.WithFields(log.Fields{
						"url":          r.repoURL,
						"suite":        suite,
						"component":    component,
						"architecture": architecture,
					}).Info("processing repository ...")

					url := fmt.Sprintf("%s/dists/%s/%s/binary-%s/Packages.gz", r.repoURL, suite, component, architecture)
					pkgs := Packages{}
					data, err := fetchMetadata(ctx, url, &pkgs)
					if err != nil {
						return err
					}

					checksum, err := sha256FromBytes(data)
					if err != nil {
						return err
					}

					for _, filename := range []string{
						fmt.Sprintf("dists/%s/%s/binary-%s/Packages.gz", suite, component, architecture),
						fmt.Sprintf("dists/%s/%s/Contents-%s.gz", suite, component, architecture),
						fmt.Sprintf("dists/%s/%s/binary-%s/by-hash/SHA256/%s", suite, component, architecture, checksum),
					} {
						if err := handler(ctx, source.Object{
							Path: filename,
							Contents: func(ctx context.Context) (io.Reader, error) {
								return bytes.NewReader(data), nil
							},
							SHA256:   checksum,
							Size:     uint64(len(data)),
							MimeType: http.DetectContentType(data),
						}); err != nil {
							return err
						}
					}

					for _, filename := range []string{
						fmt.Sprintf("dists/%s/%s/binary-%s/Release", suite, component, architecture),
						fmt.Sprintf("dists/%s/%s/binary-%s/Release.gpg", suite, component, architecture),
						fmt.Sprintf("dists/%s/%s/binary-%s/InRelease", suite, component, architecture),
					} {
						data, err := getFile(ctx, fmt.Sprintf("%s/%s", r.repoURL, filename))
						if err != nil {
							return err
						}

						checksum, err := sha256FromBytes(data)
						if err != nil {
							return err
						}

						if err := handler(ctx, source.Object{
							Path: filename,
							Contents: func(ctx context.Context) (io.Reader, error) {
								return bytes.NewReader(data), nil
							},
							SHA256:   checksum,
							Size:     uint64(len(data)),
							MimeType: http.DetectContentType(data),
						}); err != nil {
							return err
						}
					}

					for _, pkg := range pkgs {
						if err := func(pkg Package) error {
							lb := lazyblob.New(r.repoURL+"/"+pkg.Filename, os.TempDir(), uint64(pkg.Size))
							defer func() {
								if err := lb.Close(); err != nil {
									log.Warnf("error removing scratch data: %s", err)
								}
							}()

							if err := handler(ctx, source.Object{
								Path:     pkg.Filename,
								Contents: lb.Reader,
								SHA256:   pkg.SHA256Sum,
								Size:     uint64(pkg.Size),
								MimeType: detectMimeTypeByFilename(pkg.Filename),
							}); err != nil {
								return err
							}

							return nil
						}(pkg); err != nil {
							return err
						}
					}
					return nil
				}(component, architecture); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
