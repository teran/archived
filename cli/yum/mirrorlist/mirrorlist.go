package mirrorlist

import (
	"bufio"
	"context"
	"math/rand"
	"net/http"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

type Mirrorlist interface {
	URL(SelectMode) string
}

type mirrorlist struct {
	mirrors []string
	randFn  func(max int) int
}

type SelectMode string

const (
	SelectModeFirstOnly SelectMode = "firstOnly"
	SelectModeRandom    SelectMode = "random"
)

var ErrEmptyMirrorlist = errors.New("Empty mirrorlist received: no mirrors available")

func New(ctx context.Context, mirrorlistURL string) (Mirrorlist, error) {
	return newWithRandom(ctx, mirrorlistURL, rand.Intn)
}

func newWithRandom(ctx context.Context, mirrorlistURL string, randFn func(max int) int) (Mirrorlist, error) {
	mirrors, err := getMirrors(ctx, mirrorlistURL)
	if err != nil {
		return nil, err
	}

	if len(mirrors) < 1 {
		return nil, ErrEmptyMirrorlist
	}

	return &mirrorlist{
		mirrors: mirrors,
		randFn:  randFn,
	}, nil
}

func (m *mirrorlist) URL(selectMode SelectMode) string {
	switch selectMode {
	case SelectModeRandom:
		return m.mirrors[m.randFn(len(m.mirrors))]
	case SelectModeFirstOnly:
		return m.mirrors[0]
	default:
		// Just a safer option without any harm to the consumer
		log.Warnf("unexpected mode provided: %s; falling back to firstOnly", string(selectMode))
		return m.mirrors[0]
	}
}

func getMirrors(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	mirrors := []string{}
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") || strings.HasPrefix(line, "ftp://") {
			mirrors = append(mirrors, line)
		}
	}

	return mirrors, nil
}
