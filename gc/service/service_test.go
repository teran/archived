package service

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	gtm "github.com/teran/go-time"

	"github.com/teran/archived/models"
	repoMock "github.com/teran/archived/repositories/metadata/mock"
)

const defaultNamespace = "default"

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestAll() {
	s.tp.On("Now").Return("2024-08-01T10:11:12Z").Once()

	call0 := s.repoMock.On("ListNamespaces").Return([]string{defaultNamespace}, nil).Once()

	call1 := s.repoMock.On("ListContainers", defaultNamespace).Return([]string{"container1"}, nil).Once().NotBefore(call0)

	call2 := s.repoMock.On("ListUnpublishedVersionsByContainer", defaultNamespace, "container1").Return([]models.Version{
		{
			Name:      "version1",
			CreatedAt: time.Date(2024, 7, 31, 10, 1, 1, 0, time.UTC),
		},
		{
			Name:      "version2",
			CreatedAt: time.Date(2024, 8, 1, 10, 1, 1, 0, time.UTC),
		},
	}, nil).Once().NotBefore(call1)
	call3 := s.repoMock.On("ListObjects", defaultNamespace, "container1", "version1", uint64(0), uint64(1000)).Return(uint64(3), []string{"obj1", "obj2", "obj3"}, nil).Once().NotBefore(call2)
	call4 := s.repoMock.On("DeleteObject", defaultNamespace, "container1", "version1", []string{"obj1", "obj2", "obj3"}).Return(nil).Once().NotBefore(call3)
	call5 := s.repoMock.On("ListObjects", defaultNamespace, "container1", "version1", uint64(0), uint64(1000)).Return(uint64(0), []string{}, nil).Once().NotBefore(call4)
	_ = s.repoMock.On("DeleteVersion", defaultNamespace, "container1", "version1").Return(nil).Once().NotBefore(call5)

	err := s.svc.Run(s.ctx)
	s.Require().NoError(err)
}

// Definitions ...
type serviceTestSuite struct {
	suite.Suite

	ctx      context.Context
	svc      Service
	repoMock *repoMock.Mock
	tp       *gtm.TimeNowMock
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.TODO()

	s.repoMock = repoMock.New()

	s.tp = gtm.NewTimeNowMock()

	var err error
	s.svc, err = New(&Config{
		MdRepo:                   s.repoMock,
		DryRun:                   false,
		UnpublishedVersionMaxAge: 10 * time.Hour,
		TimeNowFunc:              s.tp.Now,
	})
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TearDownTest() {
	s.repoMock.AssertExpectations(s.T())
	s.tp.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
