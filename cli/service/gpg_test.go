package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"
	ptr "github.com/teran/go-ptr"

	"github.com/teran/archived/repositories/blob/mock"
)

func TestGetGPGKey(t *testing.T) {
	ctx := context.TODO()

	m := &testHandlerMock{}
	defer m.AssertExpectations(t)

	data, err := os.ReadFile("./testdata/gpg/somekey.gpg")
	if err != nil {
		t.Fatal(err)
	}

	m.On("StaticFile", "/").Return(http.StatusOK, "text/plain", data).Once()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/*", m.StaticFile)

	srv := httptest.NewServer(e)
	defer srv.Close()

	type testCase struct {
		name         string
		url          string
		expKeyIDs    []uint64
		expErrorText *string
	}

	tcs := []testCase{
		{
			name:      "read from file",
			url:       "file://./testdata/gpg/somekey.gpg",
			expKeyIDs: []uint64{11127004574349501168},
		},
		{
			name:      "read form HTTP URL",
			url:       srv.URL,
			expKeyIDs: []uint64{11127004574349501168},
		},
		{
			name: "incorrect scheme",
			expErrorText: ptr.String(
				"unexpected public key file path format. Please use file:///path/to/file.gpg or http://example.com/file.gpg"),
		},
		{
			name: "unknown scheme",
			url:  "ftp://example.com/file.gpg",
			expErrorText: ptr.String(
				"unsupported key file access scheme: `ftp`"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			keys, err := getGPGKey(ctx, tc.url)
			if tc.expErrorText != nil {
				r.Error(err)
				r.Equal(*tc.expErrorText, err.Error())
			} else {
				r.NoError(err)
				r.Equal(tc.expKeyIDs, func(el openpgp.EntityList) []uint64 {
					keyIDs := []uint64{}
					for _, key := range el {
						keyIDs = append(keyIDs, key.PrimaryKey.KeyId)
					}
					return keyIDs
				}(keys))
			}
		})
	}
}

type testHandlerMock struct {
	mock.Mock
}

func (m *testHandlerMock) StaticFile(c echo.Context) error {
	args := m.Called(c.Request().RequestURI)
	return c.Blob(args.Int(0), args.String(1), args.Get(2).([]byte))
}
