package service

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/teran/archived/models"
	repoMock "github.com/teran/archived/repositories/metadata/mock"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestAll() {
	call1 := s.repoMock.On("ListContainers").Return([]string{"container1"}, nil).Once()

	call2 := s.repoMock.On("ListUnpublishedVersionsByContainer", "container1").Return([]models.Version{
		{Name: "version1"},
	}, nil).Once().NotBefore(call1)
	call3 := s.repoMock.On("ListObjects", "container1", "version1", uint64(0), uint64(1000)).Return(uint64(3), []string{"obj1", "obj2", "obj3"}, nil).Once().NotBefore(call2)
	call4 := s.repoMock.On("DeleteObject", "container1", "version1", []string{"obj1", "obj2", "obj3"}).Return(nil).Once().NotBefore(call3)
	call5 := s.repoMock.On("ListObjects", "container1", "version1", uint64(0), uint64(1000)).Return(uint64(0), []string{}, nil).Once().NotBefore(call4)
	_ = s.repoMock.On("DeleteVersion", "container1", "version1").Return(nil).Once().NotBefore(call5)

	err := s.svc.Run(s.ctx)
	s.Require().NoError(err)
}

// Definitions ...
type serviceTestSuite struct {
	suite.Suite

	ctx      context.Context
	svc      Service
	repoMock *repoMock.Mock
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.TODO()

	s.repoMock = repoMock.New()

	var err error
	s.svc, err = New(&Config{
		MdRepo:                   s.repoMock,
		DryRun:                   false,
		UnpublishedVersionMaxAge: 10 * time.Hour,
	})
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TearDownTest() {
	s.repoMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
