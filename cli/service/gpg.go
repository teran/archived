package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"
)

func getGPGKey(ctx context.Context, filepath string) (openpgp.EntityList, error) {
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
		defer fp.Close()

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
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading public key data")
		}
	default:
		return nil, errors.Errorf("unsupported key file access scheme: `%s`", p[0])
	}

	return openpgp.ReadArmoredKeyRing(bytes.NewReader(data))
}
