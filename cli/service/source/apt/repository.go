package apt

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
	debian "pault.ag/go/debian/control"
)

var (
	errNoMetadataFileFound = errors.New("no metadata file found")
	errChecksumMismatch    = errors.New("checksum mismatch")
	errNoDecompressorFound = errors.New("decompressor is not found")
)

type metadataFile struct {
	contents []byte
	sha256   string
	size     uint64
}

type metadata map[string]metadataFile

const (
	gzipExtension = ".gz"
	xzExtension   = ".xz"
)

func (m metadata) getSuiteRelease(suite string) ([]byte, error) {
	for _, fn := range []string{
		fmt.Sprintf("dists/%s/Release", suite),
		fmt.Sprintf("dists/%s/Release.gz", suite),
		fmt.Sprintf("dists/%s/Release.xz", suite),
	} {
		if mdf, ok := m[fn]; ok {
			return getUncompressedReader(fn, mdf.contents)
		}
	}
	return nil, errNoMetadataFileFound
}

func getUncompressedReader(filename string, in []byte) ([]byte, error) {
	switch filepath.Ext(filename) {
	case "":
		return in, nil
	case gzipExtension:
		rd, err := gzip.NewReader(bytes.NewReader(in))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing gzip reader")
		}
		defer rd.Close()

		data, err := io.ReadAll(rd)
		if err != nil {
			return nil, errors.Wrap(err, "error reading decompressed stream")
		}

		return data, nil
	case xzExtension:
		rd, err := xz.NewReader(bytes.NewReader(in))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing xz reader")
		}

		data, err := io.ReadAll(rd)
		if err != nil {
			return nil, errors.Wrap(err, "error reading decompressed stream")
		}

		return data, nil
	}
	return nil, errNoDecompressorFound
}

type aptRepository interface{}

type aptRepo struct {
	baseURL           string
	enforceSignatures bool

	metadata metadata
	mutex    *sync.RWMutex
}

func newRepo(baseURL string, enforceSignatures bool) aptRepository {
	return &aptRepo{
		baseURL:           baseURL,
		enforceSignatures: enforceSignatures,

		metadata: make(metadata),
		mutex:    &sync.RWMutex{},
	}
}

func (r *aptRepo) fetchSuiteMetadata(ctx context.Context, suite string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, fn := range []string{
		fmt.Sprintf("dists/%s/Release", suite),
		fmt.Sprintf("dists/%s/Release.gz", suite),
		fmt.Sprintf("dists/%s/Release.xz", suite),
		fmt.Sprintf("dists/%s/Release.gpg", suite),
		fmt.Sprintf("dists/%s/ChangeLog", suite),
		fmt.Sprintf("dists/%s/InRelease", suite),
	} {
		release, err := getFile(ctx, fmt.Sprintf("%s/%s", r.baseURL, fn))
		if err != nil {
			log.WithFields(log.Fields{
				"suite": suite,
				"file":  fn,
			}).Warn("file doesn't exist")
			continue
		}

		checksum, err := sha256FromBytes(release)
		if err != nil {
			return errors.Wrap(err, "error calculating checksum")
		}

		r.metadata[fn] = metadataFile{
			contents: release,
			sha256:   checksum,
			size:     uint64(len(release)),
		}
	}

	// FIXME: enforceSignatures

	data, err := r.metadata.getSuiteRelease(suite)
	if err != nil {
		return errors.Wrap(err, "error getting Release data")
	}

	v := RepositoryRelease{}
	if err := debian.Unmarshal(v, bytes.NewReader(data)); err != nil {
		return errors.Wrap(err, "error unmarshaling suite Release file")
	}

	for _, md := range v.SHA256Sum {
		data, err := getFile(ctx, md.Filename)
		if err != nil {
			return errors.Wrapf(err, "error fetching file %s", md.Filename)
		}

		checksum, err := sha256FromBytes(data)
		if err != nil {
			return errors.Wrap(err, "error calculating checksum")
		}

		if checksum != md.Hash {
			return errors.Wrap(errChecksumMismatch, md.Filename)
		}

		r.metadata[md.Filename] = metadataFile{
			contents: data,
			sha256:   checksum,
			size:     uint64(len(data)),
		}
	}

	return nil
}

func (r *aptRepo) listComponents(suite string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	v := RepositoryRelease{}
	if err := debian.Unmarshal(v, bytes.NewReader(r.metadata[fmt.Sprintf("dists/%s/Release", suite)].contents)); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling suite Release file")
	}

	return v.Components, nil
}

func (r *aptRepo) listSuiteArchitectures(suite string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	v := RepositoryRelease{}
	if err := debian.Unmarshal(v, bytes.NewReader(r.metadata[fmt.Sprintf("dists/%s/Release", suite)].contents)); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling suite Release file")
	}

	return v.Architectures, nil
}

func (r *aptRepo) getPackages(component, architecture string) (Packages, error) {
	for _, fn := range []string{
		fmt.Sprintf("%s/binary-%s/Packages", component, architecture),
		fmt.Sprintf("%s/binary-%s/Packages.gz", component, architecture),
		fmt.Sprintf("%s/binary-%s/Packages.xz", component, architecture),
	} {
		if mdf, ok := r.metadata[fn]; ok {
			data, err := getUncompressedReader(fn, mdf.contents)
			if err != nil {
				return nil, errors.Wrap(err, "error getting uncompressed reader")
			}

			v := Packages{}
			if err := debian.Unmarshal(v, bytes.NewReader(data)); err != nil {
				return nil, errors.Wrap(err, "error unmarshaling Packages file")
			}

			return v, nil
		}
	}
	return nil, errNoMetadataFileFound
}

func (r *aptRepo) getMetadata() (map[string]metadataFile, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.metadata, nil
}
