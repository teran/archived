package apt

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	debian "pault.ag/go/debian/control"
)

type metadataFile struct {
	contents []byte
	sha256   string
	size     uint64
}

type aptRepository interface{}

type aptRepo struct {
	baseURL           string
	enforceSignatures bool

	metadata map[string]metadataFile
	mutex    *sync.RWMutex
}

func newRepo(baseURL string, enforceSignatures bool) aptRepository {
	return &aptRepo{
		baseURL:           baseURL,
		enforceSignatures: enforceSignatures,

		metadata: make(map[string]metadataFile),
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
	} {
		release, err := getFile(ctx, fmt.Sprintf("%s/%s", r.baseURL, fn))
		if err != nil {
			return errors.Wrap(err, "error fetching suite Release file")
		}

		releaseChecksum, err := sha256FromBytes(release)
		if err != nil {
			return errors.Wrap(err, "error calculating Release file checksum")
		}

		r.metadata[fn] = metadataFile{
			contents: release,
			sha256:   releaseChecksum,
			size:     uint64(len(release)),
		}
	}

	// FIXME: enforceSignatures

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

func (r *aptRepo) getMetadata() (map[string]metadataFile, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.metadata, nil
}
