package apt

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
	debian "pault.ag/go/debian/control"

	"github.com/teran/archived/cli/lazyblob"
	"github.com/teran/archived/cli/service/source"
)

var _ source.Source = (*repository)(nil)

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

	for _, suite := range r.suites {
		log.WithFields(log.Fields{
			"suite": suite,
		}).Trace("processing suite ...")

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

func fetchMetadata[T any](ctx context.Context, url string, v T) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rd io.Reader
	switch filepath.Ext(url) {
	case ".gz":
		rd, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case ".xz":
		rd, err = xz.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	default:
		rd = resp.Body
	}

	if err := debian.Unmarshal(v, rd); err != nil {
		return nil, err
	}

	return data, nil
}

func sha256FromBytes(in []byte) (string, error) {
	hasher := sha256.New()
	n, err := hasher.Write(in)
	if err != nil {
		return "", err
	}

	if n != len(in) {
		return "", io.ErrShortWrite
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func detectMimeTypeByFilename(path string) string {
	switch filepath.Ext(path) {
	case ".deb":
		return "application/vnd.debian.binary-package"
	case ".gz":
		return "application/x-gzip"
	default:
		return "application/octet-stream"
	}
}

func getFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
