package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	ptr "github.com/teran/go-ptr"

	cacheMock "github.com/teran/archived/cli/service/stat_cache/mock"
)

const defaultNamespace = "default"

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestCreateNamespace() {
	s.cliMock.On("CreateNamespace", "test-namespace").Return(nil).Once()

	fn := s.svc.CreateNamespace("test-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestRenameNamespace() {
	s.cliMock.On("RenameNamespace", "old-name", "new-name").Return(nil).Once()

	fn := s.svc.RenameNamespace("old-name", "new-name")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListNamespaces() {
	s.cliMock.On("ListNamespaces").Return([]string{"namespace1", "namespace2"}, nil).Once()

	fn := s.svc.ListNamespaces()
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteNamespace() {
	s.cliMock.On("DeleteNamespace", "test-namespace").Return(nil).Once()

	fn := s.svc.DeleteNamespace("test-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateContainer() {
	s.cliMock.On("CreateContainer", defaultNamespace, "test-container").Return(nil).Once()

	fn := s.svc.CreateContainer(defaultNamespace, "test-container")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestMoveContainer() {
	s.cliMock.On("MoveContainer", defaultNamespace, "test-container", "new-namespace").Return(nil).Once()

	fn := s.svc.MoveContainer(defaultNamespace, "test-container", "new-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestRenameContainer() {
	s.cliMock.On("RenameContainer", defaultNamespace, "old-name", "new-name").Return(nil).Once()

	fn := s.svc.RenameContainer(defaultNamespace, "old-name", "new-name")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListContainers() {
	s.cliMock.On("ListContainers", defaultNamespace).Return([]string{"container1", "container2"}, nil).Once()

	fn := s.svc.ListContainers(defaultNamespace)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteContainer() {
	s.cliMock.On("DeleteContainer", defaultNamespace, "test-container1").Return(nil).Once()

	fn := s.svc.DeleteContainer(defaultNamespace, "test-container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersion() {
	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", false, nil, nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionAndPublish() {
	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil).Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, nil, nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionAndPublishWithEmptyPath() {
	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil).Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String(""), nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromDirAndPublish() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String("testdata/repo"), nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublish() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6156" {
			return c.NoContent(http.StatusLengthRequired)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusUnsupportedMediaType)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil, nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053-primary.xml.gz", "12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053", uint64(4550)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5-primary.sqlite.bz2", "3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5", uint64(9036)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd-filelists.xml.gz", "f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd", uint64(3000)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52-filelists.sqlite.bz2", "a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52", uint64(5554)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f-other.xml.gz", "3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f", uint64(2848)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32-other.sqlite.bz2", "27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32", uint64(4351)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/repomd.xml", "904c00f4c838f67d1c79113d7996840add665d513889b112bb715776607c151c", uint64(3078)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-1.x86_64.rpm", "49fd5f21e3d489e500eba418207d62bc241d9f512a23b14fecb8a7777ee01bc6", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-2.x86_64.rpm", "08b032391b745436fd9b2a19b3f74889a5965c24d27d1d818a0d49b66ac4f47a", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-3.x86_64.rpm", "d9c6b377f3484f5a6312164df290186d9ed5536a8e3e67ef1c7ae8c9b956794b", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-1.x86_64.rpm", "af60a18cd0517920acac3ff737e5595c4b2dc779dd9205f6731eeba54265dfd7", uint64(6740)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-2.x86_64.rpm", "6dc90f4258183bb61baa2a0a588732fa5db0fd165d1b7bf02ff0198fca93efdb", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-3.x86_64.rpm", "a68d0510bec578428402a029eac55e34e8784e96556e9f2d1c9424911c2a489f", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-1.x86_64.rpm", "88307ad55751656bb96132d8d448aa78ebafb28120634695f8207c8460ad7dff", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-2.x86_64.rpm", "e5dd6aac17915a9ada1b39e4efcb03e4ee3e6998ff910c0584ccb11f97721632", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-3.x86_64.rpm", "d216d06a4ff1569a98ff981469bbf2765300587f69dbf378a150c0cf8dfb4795", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-1.x86_64.rpm", "68170e526f756eeb06d33d77332000a67e8eb31eb004f28de509ed8aae72f8a6", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-2.x86_64.rpm", "f75513d7fc853e3f8d4c832700151dcb04d88388f367c4a29a0c702988fd9c80", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-3.x86_64.rpm", "ec036e47e7e711cf26b126c103e25bf0191a67bfc44089f18ee6c3ade8a334d1", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-1.x86_64.rpm", "e7de03570c76ec13cceff7bca33cec6fcff45e3c4c83e088b06764825542e78b", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-2.x86_64.rpm", "c9731ea0936b1c4debc4ebae7935687c4a5a5f9b538ab795c903adbce6ba5b70", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-3.x86_64.rpm", "5942505f4082335d8cde4e03bb9bb9080e86efb6a24dce6167a71abbe418a1d2", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-1.x86_64.rpm", "006ef00d887654372886d295c1dc03eff7b18c611c1d35c4081456300bf99d15", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-2.x86_64.rpm", "ab195eaac0cc29c13c5920e0bdb8f2dedcc20ca10861775d6a0ccce8bb81546c", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-3.x86_64.rpm", "f2149561e84c54c429869ec051b18b3c59a92313164fc2b678c08b182711e339", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-1.x86_64.rpm", "234b5d4a4ea3cd320292c374efdb2ca51876fac43f902d3816a8b04bd155a5c1", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-2.x86_64.rpm", "2665d47c785688562748b0d7c0637a2c9ff052cb12acf09a2b64d0da699843f2", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-3.x86_64.rpm", "84c25528dc5381c89656e52a6af6fa51c7cb9faa69cec3b5c196ffd4ba037c80", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-1.x86_64.rpm", "6b3edf4d9a1194b0b07c660281378375d258f25e1e74c7852f9fb209c0b1de4f", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-2.x86_64.rpm", "8c9d92872e4c3c462203682c63c79d61c00b398f84a4adf66f2da08e4c416a4b", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-3.x86_64.rpm", "845b31dbb4d80b13ace39ab2da1e19d68727d43de7b01fd5cf5c10e633f7707e", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-1.x86_64.rpm", "1ed2f5e08a1f57c68199c57271404945762bcba36d99c2c03a45d60ad8e53a75", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-2.x86_64.rpm", "dde30ee7035800604434e3b0928934c56fc382748cb3dc6a87ee7605a15af9a3", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-3.x86_64.rpm", "1fec4e8fcd788227fd836960f129b7c5384bf416aa78bdb7923dfdc431b208bf", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-1.x86_64.rpm", "0ba29dded6b136adb548396295f46ff4fa827d416cdfa25129ddaef5be2f2d96", uint64(6740)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-2.x86_64.rpm", "192d106bc5c19b9e8651dbd255f5f0a240b781b85a2af0ce58d1c5f8670158b6", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-3.x86_64.rpm", "ae2f90464235b2abeccf9c2607c51bab0b8a4f3afd6531026d62d54300b2e093", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-1.src.rpm", "5906d8401381f428c28074563ed082425074cf4737ef38ca1bf21c3261aabd76", uint64(6123)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-2.src.rpm", "fb69e232a677646c0375e9cf999e5c8493368edcafd0e5fe6f901a7425f7d68e", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-3.src.rpm", "cf28031ec6d8ddc146b1f2ec37e4b7bbd7f4f751ae9b7730f709c43d0292bb79", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-1.src.rpm", "7f86347249889174a4bcfa2d20aec471e44275be06484f45a2d1afa4f61d8895", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-2.src.rpm", "151a800766bfb51b8f414878367f64028f3693dc323a01de2ccc10d96813f302", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-3.src.rpm", "370e830462c33db0cc11dcaa7f539773a651c1dde980bd1e92db601d85156419", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-1.src.rpm", "802c968cd29794edf721fcbc5924a68606df9a0648d89c648481bf77ff89c53c", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-2.src.rpm", "f26e8390a958c0fd0a15c07cf9ef2f93485b90ec701888013f1453b0946c7c47", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-3.src.rpm", "07a1e8b441c8e329cb07567bdbd845d833da6721dfc166c3510469f313ebd652", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-1.src.rpm", "3a557ebe499975f3c7b95e1b39ca3ff10369b51dec045d8f882a0cb587815db4", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-2.src.rpm", "4ad8107d6a8a9f32c3c2f4971756467ee536979cb246e53a9d238c824df665ea", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-3.src.rpm", "14b797a378b75c5c6793400f2d2f14b98099ba32b95828686ca7b14c8475d889", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-1.src.rpm", "7b981767615dd5c739f027bcc17aa91343d36aa5040e59f7614e401a35b2e703", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-2.src.rpm", "3f55e2dd76104d19baf4148342c0cdc5b0230879cca6c5b13c07f24a25d79a1d", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-3.src.rpm", "80ea7ece7614339e906e80920a5a5d5bf0881f3e9a0a02ab2831036e67ea4152", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-1.src.rpm", "182367519219c7f56c1213350372333219fa0d2f858072f524734308949daf61", uint64(6122)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-2.src.rpm", "904021dcfd97e54d3c5cbaa7742af844cc5c48adb4af4ba88495db30afd629d9", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-3.src.rpm", "7b74718b1f01b561f66d04706327269bf67981db6743c46858fe03e2868c6ef5", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-1.src.rpm", "96e669c98d8d79d0acc77f1d9b3246e6baabf2f49a7591d26ba69740b80574ee", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-2.src.rpm", "798464bf773513c123183fe00aa350459ec31568e2ee1895d8fef4d6dfbb2e28", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-3.src.rpm", "d9e1eaa79d2b80349186b3ec51a5331ca400994d9ba0a5fbcd6992fdf2234fcb", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-1.src.rpm", "8f76905b96b5476485fca223804a9900c3351f0b61e30f862ebc7e48f1232fca", uint64(6123)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-2.src.rpm", "263e504fd0be3d6903e03e4b5d3c820c6f06e474253d63bf7951223c3080cc17", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-3.src.rpm", "c04e4e225556951502e2ce533ac25d26decf63a768013c14bcec32dc60b88837", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-1.src.rpm", "6b0bb85fe30ae6e406a1dddb2af7126294e24fae64e79be67c426e94f90720ee", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-2.src.rpm", "cc3fd1c43ea62332e534e0f5d6a7df7f1c808725116109f33fe38453682d5ee7", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-3.src.rpm", "d5ccf0943d89947bb2f8b3ed8466d60e452f1b0daa881d5d3fe2c8584778883b", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-1.src.rpm", "ed76da35a9b4f4ea4e3828f6a978016cc6b4b9c0cd0548f2d4cbd6a695b1f6e9", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-2.src.rpm", "26dc22dbbbd6569c6d41d7c37b6e8b3d90ec24c76c075bfe8b0016de64bb2029", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-3.src.rpm", "869bf9b01e2ea53e02bb8666ca1118e2c9e5cef2598e9a1b1f67125531b3bd8e", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String(""), ptr.String(srv.URL), ptr.String(""))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublishSHA1() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo-sha1")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6156" {
			return c.NoContent(http.StatusLengthRequired)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusUnsupportedMediaType)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/80779e2ab55e25a77124d370de1d08deae8f1cc6-primary.xml.gz", "1c07f3f3f0e6d09972c1d7852d1dbc9715d6fbdceee66c50e8356d1e69502d3b", uint64(688)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/e7a8a53e7398f6c22894718ea227fea60f2b78ba-primary.sqlite.bz2", "c9b8ce03b503e29d9ec2faa2328e4f2082f0a5f71478ca6cb2f1a3ab75e676bc", uint64(1937)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/4a11e3eeb25d21b08f41e5578d702d2bea21a2e7-filelists.xml.gz", "b56801c0a86f9a0136953e8c8e59cd35c1f18fc41e70ba8fcdcccfee068dfc8a", uint64(282)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/c66ce2caa41ed83879f9b3dd9f40e61c65af499e-filelists.sqlite.bz2", "59bd3edd4edacac87e5e15494698f34a7f52277691635f927c185e92a681d9ee", uint64(787)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/fdedb6ce109127d52228d01b0239010ddca14c8f-other.xml.gz", "56e566dfc63b0a7056b21cec661717a411f68cf98747d9a719557bce3a8ac41a", uint64(247)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/b31561a27d014d35b59b27c27859bb1c17ac573e-other.sqlite.bz2", "7eec446e0036d356d8e5694047d9fdb6af00f2fc62993b854232830cf9dbcff8", uint64(669)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/repomd.xml", "9f18801e8532f631e308a130a347f66eb3900d054df1d66dff53a69aa5b9e7d3", uint64(2601)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "Packages/testpkg-1-1.src.rpm", "684303227d799ffe1f0b39e030a12ad249931a11ec1690e2079f981cc16d8c52", uint64(6156)).
		Return(ptr.String(uploadSrv.URL+"/upload"), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "Packages/testpkg-1-1.x86_64.rpm", "d9ae5e56ea38d2ac470f320cade63663dae6ab8b8e1630b2fd5a3c607f45e2ee", uint64(6722)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String(""), ptr.String(srv.URL), ptr.String(""))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublishGPGNoSignature() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6156" {
			return c.NoContent(http.StatusLengthRequired)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusUnsupportedMediaType)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053-primary.xml.gz", "12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053", uint64(4550)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5-primary.sqlite.bz2", "3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5", uint64(9036)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd-filelists.xml.gz", "f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd", uint64(3000)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52-filelists.sqlite.bz2", "a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52", uint64(5554)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f-other.xml.gz", "3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f", uint64(2848)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32-other.sqlite.bz2", "27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32", uint64(4351)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/repomd.xml", "904c00f4c838f67d1c79113d7996840add665d513889b112bb715776607c151c", uint64(3078)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-1.x86_64.rpm", "49fd5f21e3d489e500eba418207d62bc241d9f512a23b14fecb8a7777ee01bc6", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-2.x86_64.rpm", "08b032391b745436fd9b2a19b3f74889a5965c24d27d1d818a0d49b66ac4f47a", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg1-1-3.x86_64.rpm", "d9c6b377f3484f5a6312164df290186d9ed5536a8e3e67ef1c7ae8c9b956794b", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-1.x86_64.rpm", "af60a18cd0517920acac3ff737e5595c4b2dc779dd9205f6731eeba54265dfd7", uint64(6740)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-2.x86_64.rpm", "6dc90f4258183bb61baa2a0a588732fa5db0fd165d1b7bf02ff0198fca93efdb", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg10-1-3.x86_64.rpm", "a68d0510bec578428402a029eac55e34e8784e96556e9f2d1c9424911c2a489f", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-1.x86_64.rpm", "88307ad55751656bb96132d8d448aa78ebafb28120634695f8207c8460ad7dff", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-2.x86_64.rpm", "e5dd6aac17915a9ada1b39e4efcb03e4ee3e6998ff910c0584ccb11f97721632", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg2-1-3.x86_64.rpm", "d216d06a4ff1569a98ff981469bbf2765300587f69dbf378a150c0cf8dfb4795", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-1.x86_64.rpm", "68170e526f756eeb06d33d77332000a67e8eb31eb004f28de509ed8aae72f8a6", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-2.x86_64.rpm", "f75513d7fc853e3f8d4c832700151dcb04d88388f367c4a29a0c702988fd9c80", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg3-1-3.x86_64.rpm", "ec036e47e7e711cf26b126c103e25bf0191a67bfc44089f18ee6c3ade8a334d1", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-1.x86_64.rpm", "e7de03570c76ec13cceff7bca33cec6fcff45e3c4c83e088b06764825542e78b", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-2.x86_64.rpm", "c9731ea0936b1c4debc4ebae7935687c4a5a5f9b538ab795c903adbce6ba5b70", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg4-1-3.x86_64.rpm", "5942505f4082335d8cde4e03bb9bb9080e86efb6a24dce6167a71abbe418a1d2", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-1.x86_64.rpm", "006ef00d887654372886d295c1dc03eff7b18c611c1d35c4081456300bf99d15", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-2.x86_64.rpm", "ab195eaac0cc29c13c5920e0bdb8f2dedcc20ca10861775d6a0ccce8bb81546c", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg5-1-3.x86_64.rpm", "f2149561e84c54c429869ec051b18b3c59a92313164fc2b678c08b182711e339", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-1.x86_64.rpm", "234b5d4a4ea3cd320292c374efdb2ca51876fac43f902d3816a8b04bd155a5c1", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-2.x86_64.rpm", "2665d47c785688562748b0d7c0637a2c9ff052cb12acf09a2b64d0da699843f2", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg6-1-3.x86_64.rpm", "84c25528dc5381c89656e52a6af6fa51c7cb9faa69cec3b5c196ffd4ba037c80", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-1.x86_64.rpm", "6b3edf4d9a1194b0b07c660281378375d258f25e1e74c7852f9fb209c0b1de4f", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-2.x86_64.rpm", "8c9d92872e4c3c462203682c63c79d61c00b398f84a4adf66f2da08e4c416a4b", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg7-1-3.x86_64.rpm", "845b31dbb4d80b13ace39ab2da1e19d68727d43de7b01fd5cf5c10e633f7707e", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-1.x86_64.rpm", "1ed2f5e08a1f57c68199c57271404945762bcba36d99c2c03a45d60ad8e53a75", uint64(6741)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-2.x86_64.rpm", "dde30ee7035800604434e3b0928934c56fc382748cb3dc6a87ee7605a15af9a3", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg8-1-3.x86_64.rpm", "1fec4e8fcd788227fd836960f129b7c5384bf416aa78bdb7923dfdc431b208bf", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-1.x86_64.rpm", "0ba29dded6b136adb548396295f46ff4fa827d416cdfa25129ddaef5be2f2d96", uint64(6740)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-2.x86_64.rpm", "192d106bc5c19b9e8651dbd255f5f0a240b781b85a2af0ce58d1c5f8670158b6", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg9-1-3.x86_64.rpm", "ae2f90464235b2abeccf9c2607c51bab0b8a4f3afd6531026d62d54300b2e093", uint64(6742)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-1.src.rpm", "5906d8401381f428c28074563ed082425074cf4737ef38ca1bf21c3261aabd76", uint64(6123)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-2.src.rpm", "fb69e232a677646c0375e9cf999e5c8493368edcafd0e5fe6f901a7425f7d68e", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg1-1-3.src.rpm", "cf28031ec6d8ddc146b1f2ec37e4b7bbd7f4f751ae9b7730f709c43d0292bb79", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-1.src.rpm", "7f86347249889174a4bcfa2d20aec471e44275be06484f45a2d1afa4f61d8895", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-2.src.rpm", "151a800766bfb51b8f414878367f64028f3693dc323a01de2ccc10d96813f302", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg10-1-3.src.rpm", "370e830462c33db0cc11dcaa7f539773a651c1dde980bd1e92db601d85156419", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-1.src.rpm", "802c968cd29794edf721fcbc5924a68606df9a0648d89c648481bf77ff89c53c", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-2.src.rpm", "f26e8390a958c0fd0a15c07cf9ef2f93485b90ec701888013f1453b0946c7c47", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg2-1-3.src.rpm", "07a1e8b441c8e329cb07567bdbd845d833da6721dfc166c3510469f313ebd652", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-1.src.rpm", "3a557ebe499975f3c7b95e1b39ca3ff10369b51dec045d8f882a0cb587815db4", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-2.src.rpm", "4ad8107d6a8a9f32c3c2f4971756467ee536979cb246e53a9d238c824df665ea", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg3-1-3.src.rpm", "14b797a378b75c5c6793400f2d2f14b98099ba32b95828686ca7b14c8475d889", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-1.src.rpm", "7b981767615dd5c739f027bcc17aa91343d36aa5040e59f7614e401a35b2e703", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-2.src.rpm", "3f55e2dd76104d19baf4148342c0cdc5b0230879cca6c5b13c07f24a25d79a1d", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg4-1-3.src.rpm", "80ea7ece7614339e906e80920a5a5d5bf0881f3e9a0a02ab2831036e67ea4152", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-1.src.rpm", "182367519219c7f56c1213350372333219fa0d2f858072f524734308949daf61", uint64(6122)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-2.src.rpm", "904021dcfd97e54d3c5cbaa7742af844cc5c48adb4af4ba88495db30afd629d9", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg5-1-3.src.rpm", "7b74718b1f01b561f66d04706327269bf67981db6743c46858fe03e2868c6ef5", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-1.src.rpm", "96e669c98d8d79d0acc77f1d9b3246e6baabf2f49a7591d26ba69740b80574ee", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-2.src.rpm", "798464bf773513c123183fe00aa350459ec31568e2ee1895d8fef4d6dfbb2e28", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg6-1-3.src.rpm", "d9e1eaa79d2b80349186b3ec51a5331ca400994d9ba0a5fbcd6992fdf2234fcb", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-1.src.rpm", "8f76905b96b5476485fca223804a9900c3351f0b61e30f862ebc7e48f1232fca", uint64(6123)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-2.src.rpm", "263e504fd0be3d6903e03e4b5d3c820c6f06e474253d63bf7951223c3080cc17", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg7-1-3.src.rpm", "c04e4e225556951502e2ce533ac25d26decf63a768013c14bcec32dc60b88837", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-1.src.rpm", "6b0bb85fe30ae6e406a1dddb2af7126294e24fae64e79be67c426e94f90720ee", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-2.src.rpm", "cc3fd1c43ea62332e534e0f5d6a7df7f1c808725116109f33fe38453682d5ee7", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg8-1-3.src.rpm", "d5ccf0943d89947bb2f8b3ed8466d60e452f1b0daa881d5d3fe2c8584778883b", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-1.src.rpm", "ed76da35a9b4f4ea4e3828f6a978016cc6b4b9c0cd0548f2d4cbd6a695b1f6e9", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-2.src.rpm", "26dc22dbbbd6569c6d41d7c37b6e8b3d90ec24c76c075bfe8b0016de64bb2029", uint64(6124)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg9-1-3.src.rpm", "869bf9b01e2ea53e02bb8666ca1118e2c9e5cef2598e9a1b1f67125531b3bd8e", uint64(6125)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String(""), ptr.String(srv.URL), ptr.String("file://./testdata/gpg/somekey.gpg"))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublishGPGSigned() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo-signed")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6115" {
			return c.NoContent(http.StatusLengthRequired)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusUnsupportedMediaType)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/c99064d5c98a01f3720c735431ab1f449c56a1c4e233efd99353716d856c245f-primary.xml.gz", "c99064d5c98a01f3720c735431ab1f449c56a1c4e233efd99353716d856c245f", uint64(718)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/0586c412097e75a9420880bb8256802008e79f2cbe7d7d34cebeb55abce6ad40-primary.sqlite.bz2", "0586c412097e75a9420880bb8256802008e79f2cbe7d7d34cebeb55abce6ad40", uint64(1985)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/64f76a79439371fc632b7cac21b68f322142bc183b706332314e97d1007f8f0c-filelists.xml.gz", "64f76a79439371fc632b7cac21b68f322142bc183b706332314e97d1007f8f0c", uint64(314)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/464b3eff37b3eee86b7e4f78efcf0e8911afa496a57753ac42a67c2afbdd2d48-filelists.sqlite.bz2", "464b3eff37b3eee86b7e4f78efcf0e8911afa496a57753ac42a67c2afbdd2d48", uint64(859)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/4c25d51dbded8086515c32ce5753f7bd22d4b0d0ee9c45d3f580751fbd26e05a-other.xml.gz", "4c25d51dbded8086515c32ce5753f7bd22d4b0d0ee9c45d3f580751fbd26e05a", uint64(282)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/f16e9edf15ee11cb5a79dd9466f58bbe2a481db47cfad8f6287540beda0779f6-other.sqlite.bz2", "f16e9edf15ee11cb5a79dd9466f58bbe2a481db47cfad8f6287540beda0779f6", uint64(743)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "repodata/repomd.xml", "4d207e2d80ec3aefb6f9e08f744f547f7171c94dc451d01fa24fe5c57ffb01a0", uint64(3069)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "SRPMS/testpkg-1-1.src.rpm", "cbce80483872b31a4b92a9ef0aea11f38e0f06db301781db53ba88a365bffd8e", uint64(6115)).
		Return(ptr.String(uploadSrv.URL+"/upload"), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version_id", "RPMS/x86_64/testpkg-1-1.x86_64.rpm", "3ea740db3d27481b38231c9bd987c46bb6bdda480c60fbfcce84d7d88abf5051", uint64(6734)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, ptr.String(""), ptr.String(srv.URL), ptr.String("file://./testdata/gpg/somekey.gpg"))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.cliMock.On("DeleteVersion", defaultNamespace, "container1", "version1").Return(nil).Once()

	fn := s.svc.DeleteVersion(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListVersions() {
	s.cliMock.On("ListVersions", defaultNamespace, "container1").Return([]string{"version1", "version2", "version3"}, nil).Once()

	fn := s.svc.ListVersions(defaultNamespace, "container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version1").Return(nil).Once()

	fn := s.svc.PublishVersion(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithoutEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithCache() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithUploadURL() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		defer r.Body.Close()

		s.Require().Equal("1234\n", string(data))

		s.Require().Equal("/test-url", r.RequestURI)
		s.Require().Equal("multipart/form-data", r.Header.Get("Content-Type"))
	}))
	defer srv.Close()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(srv.URL+"/test-url"), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteObject() {
	s.cliMock.On("DeleteObject", defaultNamespace, "container1", "version1", "key1").Return(nil).Once()

	fn := s.svc.DeleteObject(defaultNamespace, "container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListObjects() {
	s.cliMock.On("ListObjects", defaultNamespace, "container1", "version1").Return([]string{"obj1", "obj2", "obj3"}, nil).Once()

	fn := s.svc.ListObjects(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestGetObjectURL() {
	s.cliMock.On("GetObjectURL", defaultNamespace, "container1", "version1", "key1").Return("https://example.com", nil).Once()

	fn := s.svc.GetObjectURL(defaultNamespace, "container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

// Definitions ...
type serviceTestSuite struct {
	suite.Suite

	ctx       context.Context
	cliMock   *protoClientMock
	cacheMock *cacheMock.Mock
	svc       Service
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.cliMock = newMock()
	s.cacheMock = cacheMock.New()

	s.svc = New(s.cliMock, s.cacheMock)
}

func (s *serviceTestSuite) TearDownTest() {
	s.cliMock.AssertExpectations(s.T())
	s.cacheMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
