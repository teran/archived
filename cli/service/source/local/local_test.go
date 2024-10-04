package local

import (
	"context"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/teran/archived/cli/service/source"
	"github.com/teran/archived/cli/service/source/mock"
	statCacheMock "github.com/teran/archived/cli/service/stat_cache/mock"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestChecksumFile(t *testing.T) {
	r := require.New(t)

	cs, err := checksumFile("testdata/checksum/test_file.txt")
	r.NoError(err)
	r.Equal("0c15e883dee85bb2f3540a47ec58f617a2547117f9096417ba5422268029f501", cs)
}

func TestSource(t *testing.T) {
	r := require.New(t)
	ctx := context.TODO()

	statCache := statCacheMock.New()
	defer statCache.AssertExpectations(t)

	statCache.On("Get", "testdata/dir/some_file1.txt").Return("", nil).Once()
	statCache.On("Put", "testdata/dir/some_file1.txt", "cb330beb8590577eb619d75183b14ac85d6b30a6777e8041d6c2d8a44888e7f1").Return(nil).Once()

	statCache.On("Get", "testdata/dir/some_file2.txt").Return("", nil).Once()
	statCache.On("Put", "testdata/dir/some_file2.txt", "e45fbded5effe3178f7ca393f0228fb6799ead901c8a5b1354d6f1c44c2a8fa7").Return(nil).Once()

	statCache.On("Get", "testdata/dir/some_dir/some_file3.txt").Return("", nil).Once()
	statCache.On("Put", "testdata/dir/some_dir/some_file3.txt", "94630c0572bba1a7dcc8e69a70ffbaaf132fc7f67f0c40416b73d53c10d9aa7a").Return(nil).Once()

	handle := new(handlerMock)
	defer handle.AssertExpectations(t)

	handle.On("Handle", "some_file1.txt", uint64(18), "cb330beb8590577eb619d75183b14ac85d6b30a6777e8041d6c2d8a44888e7f1").Return(nil).Once()
	handle.On("Handle", "some_file2.txt", uint64(18), "e45fbded5effe3178f7ca393f0228fb6799ead901c8a5b1354d6f1c44c2a8fa7").Return(nil).Once()
	handle.On("Handle", "some_dir/some_file3.txt", uint64(18), "94630c0572bba1a7dcc8e69a70ffbaaf132fc7f67f0c40416b73d53c10d9aa7a").Return(nil).Once()

	s := New("testdata/dir", statCache)

	err := s.Process(ctx, func(ctx context.Context, obj source.Object) error {
		return handle.Handle(obj)
	})
	r.NoError(err)
}

type handlerMock struct {
	mock.Mock
}

func (m *handlerMock) Handle(obj source.Object) error {
	args := m.Called(obj.Path, obj.Size, obj.SHA256)
	return args.Error(0)
}
