package apt

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
	debian "pault.ag/go/debian/control"
)

func fetchMetadata[T any](ctx context.Context, url string, v T) ([]byte, error) {
	rawData, err := getFile(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "error getting file")
	}

	log.WithFields(log.Fields{
		"length": len(rawData),
	}).Trace("control structure received into buffer")

	var rd io.Reader
	switch filepath.Ext(url) {
	case ".gz":
		rd, err = gzip.NewReader(bytes.NewReader(rawData))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing gzip reader")
		}
		defer rd.(*gzip.Reader).Close()
	case ".xz":
		rd, err = xz.NewReader(bytes.NewReader(rawData))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing xz reader")
		}
	default:
		rd = bytes.NewReader(rawData)
	}

	if err := debian.Unmarshal(v, rd); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling control structure")
	}

	return rawData, nil
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
	case ".xz":
		return "application/x-xz"
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

// func downloadFileAndVerify(ctx context.Context, file *downloadableFile) (err error) {
// 	file.contents, err = getFile(ctx, file.url)
// 	if err != nil {
// 		return errors.Wrap(err, "error getting file")
// 	}

// 	if file.withSHA256 != nil {
// 		checksum, err := sha256FromBytes(file.contents)
// 		if err != nil {
// 			return err
// 		}

// 		if *file.withSHA256 != checksum {
// 			return ErrChecksumMismatch
// 		}
// 	}

// 	if file.withGPGSignature {
// 		file.signatureContents, err = getFile(ctx, file.url+".gpg")
// 		if err != nil {
// 			return errors.Wrap(err, "error getting file signature")
// 		}

// 		// check signature
// 	}

// 	return nil
// }

// func getGPGKey(ctx context.Context, filepath string, checksum *string) (openpgp.EntityList, error) {
// 	p := strings.SplitN(filepath, "://", 2)
// 	if len(p) != 2 {
// 		return nil, errors.New("unexpected public key file path format. Please use file:///path/to/file.gpg or http://example.com/file.gpg")
// 	}

// 	var data []byte
// 	switch p[0] {
// 	case "file":
// 		fp, err := os.Open(p[1])
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error opening public key file")
// 		}
// 		defer fp.Close()

// 		data, err = io.ReadAll(fp)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error reading public key file")
// 		}
// 	case "http", "https":
// 		req, err := http.NewRequestWithContext(ctx, http.MethodGet, filepath, nil)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error constructing HTTP request object")
// 		}

// 		resp, err := http.DefaultClient.Do(req)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error performing HTTP request")
// 		}
// 		defer resp.Body.Close()

// 		data, err = io.ReadAll(resp.Body)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error reading public key data")
// 		}
// 	default:
// 		return nil, errors.Errorf("unsupported key file access scheme: `%s`", p[0])
// 	}

// 	if checksum != nil && *checksum != "" {
// 		h := sha256.New()
// 		n, err := h.Write(data)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "error calculating SHA256")
// 		}

// 		if n != len(data) {
// 			return nil, errors.Wrap(io.ErrShortWrite, "error writing data to hasher")
// 		}

// 		if *checksum != hex.EncodeToString(h.Sum(nil)) {
// 			return nil, errors.New("GPG Key checksum mismatch")
// 		}

// 		log.WithFields(log.Fields{
// 			"sha256": *checksum,
// 		}).Infof("GPG key checksum verified")
// 	}

// 	return openpgp.ReadArmoredKeyRing(bytes.NewReader(data))
// }
