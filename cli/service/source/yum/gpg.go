package yum

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func getGPGKey(ctx context.Context, filepath string, checksum *string) (openpgp.EntityList, error) {
	p := strings.SplitN(filepath, "://", 2)
	if len(p) != 2 {
		return nil, errors.New("unexpected public key file path format. Please use file:///path/to/file.gpg or http://example.com/file.gpg")
	}

	var data []byte
	switch p[0] {
	case "file":
		fp, err := os.Open(p[1])
		if err != nil {
			return nil, errors.Wrap(err, "error opening public key file")
		}
		defer func() { _ = fp.Close() }()

		data, err = io.ReadAll(fp)
		if err != nil {
			return nil, errors.Wrap(err, "error reading public key file")
		}
	case "http", "https":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, filepath, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error constructing HTTP request object")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "error performing HTTP request")
		}
		defer func() { _ = resp.Body.Close() }()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading public key data")
		}
	default:
		return nil, errors.Errorf("unsupported key file access scheme: `%s`", p[0])
	}

	if checksum != nil && *checksum != "" {
		h := sha256.New()
		n, err := h.Write(data)
		if err != nil {
			return nil, errors.Wrap(err, "error calculating SHA256")
		}

		if n != len(data) {
			return nil, errors.Wrap(io.ErrShortWrite, "error writing data to hasher")
		}

		if *checksum != hex.EncodeToString(h.Sum(nil)) {
			return nil, errors.New("GPG Key checksum mismatch")
		}

		log.WithFields(log.Fields{
			"sha256": *checksum,
		}).Infof("GPG key checksum verified")
	}

	return openpgp.ReadArmoredKeyRing(bytes.NewReader(data))
}
