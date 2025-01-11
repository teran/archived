package apt

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *aptSourceTestSuite) TestUnsignedRepo() {}

// Definitions ...
type aptSourceTestSuite struct {
	suite.Suite
}

func TestAptSourceTestSuite(t *testing.T) {
	suite.Run(t, &aptSourceTestSuite{})
}
