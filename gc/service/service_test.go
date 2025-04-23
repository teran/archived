package service

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	gtm "github.com/teran/go-collection/time/mock"

	repoMock "github.com/teran/archived/repositories/metadata/mock"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestDeleteUnpublishedExpiredVersions() {
	s.repoMock.On("DeleteExpiredVersionsWithObjects", 10*time.Hour).Return(nil).Once()

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
		UnpublishedVersionMaxAge: 10 * time.Hour,
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
