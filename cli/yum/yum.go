package yum

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/xml"
	"hash"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/yum/models"
)

var (
	_ YumRepo = (*yumRepo)(nil)

	ErrFileNotFound             = errors.New("file not found")
	ErrChecksumMismatch         = errors.New("checksum mismatch")
	ErrNotSupportedChecksumAlgo = errors.New("not supported checksum algorithm is in use")

	hashFunctionsByName map[string]func() hash.Hash = map[string]func() hash.Hash{
		"md5":    md5.New,
		"sha":    sha1.New,
		"sha256": sha256.New,
		"sha512": sha512.New,
	}
)

type YumRepo interface {
	Packages(ctx context.Context) ([]models.Package, error)
	Metadata() map[string][]byte
}

type yumRepo struct {
	mutex    *sync.RWMutex
	url      string
	metadata map[string][]byte
}

func New(url string) *yumRepo {
	return &yumRepo{
		url:      strings.TrimSuffix(url, "/"),
		metadata: make(map[string][]byte),
		mutex:    &sync.RWMutex{},
	}
}

func (y *yumRepo) Packages(ctx context.Context) ([]models.Package, error) {
	rd, err := fetch(ctx, y.url+"/repodata/repomd.xml")
	if err != nil {
		return nil, err
	}
	defer rd.Close()

	y.mutex.Lock()
	defer y.mutex.Unlock()
	y.metadata["repodata/repomd.xml"], err = io.ReadAll(rd)
	if err != nil {
		return nil, errors.Wrap(err, "error reading repomd.xml")
	}

	repomd := models.RepoMD{}
	if err := xml.Unmarshal(y.metadata["repodata/repomd.xml"], &repomd); err != nil {
		return nil, errors.Wrap(err, "error decoding repomd XML")
	}

	if err := y.fetchRepoMetadata(ctx, repomd); err != nil {
		return nil, errors.Wrap(err, "error fetching repository metadata")
	}

	primary, err := repomd.GetPrimary()
	if err != nil {
		return nil, err
	}

	return y.fetchPackageIndex(ctx, primary.Location, primary.Checksum)
}

func (y *yumRepo) Metadata() map[string][]byte {
	y.mutex.RLock()
	defer y.mutex.RUnlock()

	out := map[string][]byte{}
	for k, v := range y.metadata {
		out[k] = append(out[k], v...)
	}

	return out
}

func (y *yumRepo) fetchRepoMetadata(ctx context.Context, repomd models.RepoMD) error {
	for _, md := range repomd.Data {
		filename := strings.TrimPrefix(md.Location.Href, "/")

		rd, err := fetch(ctx, y.url+"/"+filename)
		if err != nil {
			return err
		}
		defer rd.Close()

		data, err := io.ReadAll(rd)
		if err != nil {
			return errors.Wrap(err, "error reading file")
		}

		y.metadata[filename] = append(y.metadata[filename], data...)
	}

	return nil
}

func (y *yumRepo) fetchPackageIndex(ctx context.Context, href models.RepoMDDataLocation, checksum models.RepoMDDataChecksum) ([]models.Package, error) {
	log.Tracef("primary index url: %s", href.Href)

	indexFileName := strings.TrimPrefix(href.Href, "/")
	rd, err := fetch(ctx, y.url+"/"+indexFileName)
	if err != nil {
		return nil, err
	}
	defer rd.Close()

	hfn, err := hasherByName(checksum.Type)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	hasher := hfn()
	trd := io.TeeReader(io.TeeReader(rd, buf), hasher)

	if strings.HasSuffix(href.Href, ".gz") {
		log.Tracef(".gz extension detected. Wrapping ...")
		trd, err = gzip.NewReader(trd)
		if err != nil {
			return nil, errors.Wrap(err, "error creating gzip decoder")
		}
	}

	y.metadata[indexFileName] = buf.Bytes()

	primaryMD := models.PrimaryMD{}
	if err := xml.NewDecoder(trd).Decode(&primaryMD); err != nil {
		return nil, errors.Wrap(err, "error decoding XML")
	}

	if hex.EncodeToString(hasher.Sum(nil)) != checksum.Text {
		return nil, ErrChecksumMismatch
	}

	packages := []models.Package{}
	for _, pkg := range primaryMD.Package {
		packages = append(packages, models.Package{
			Name:         pkg.Location.Href,
			Checksum:     pkg.Checksum.Text,
			ChecksumType: pkg.Checksum.Type,
			Size:         pkg.Size.Package,
		})
	}

	return packages, nil
}

func fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	log.Tracef("requesting `%s` ...", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, ErrFileNotFound
	}

	return resp.Body, nil
}

func hasherByName(algo string) (func() hash.Hash, error) {
	if h, ok := hashFunctionsByName[algo]; ok {
		return h, nil
	}
	return nil, errors.Wrapf(ErrNotSupportedChecksumAlgo, "requested algo is `%s`", algo)
}
