package yum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/teran/archived/cli/service/source/yum/yum_repo/models"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestPackages(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	packages, err := repo.Packages(context.Background())
	r.NoError(err)
	r.Equal([]models.Package{
		{
			Name:         "SRPMS/testpkg1-1-1.src.rpm",
			Checksum:     "5906d8401381f428c28074563ed082425074cf4737ef38ca1bf21c3261aabd76",
			ChecksumType: "sha256",
			Size:         6123,
		},
		{
			Name:         "RPMS/x86_64/testpkg1-1-1.x86_64.rpm",
			Checksum:     "49fd5f21e3d489e500eba418207d62bc241d9f512a23b14fecb8a7777ee01bc6",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg1-1-2.src.rpm",
			Checksum:     "fb69e232a677646c0375e9cf999e5c8493368edcafd0e5fe6f901a7425f7d68e",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg1-1-2.x86_64.rpm",
			Checksum:     "08b032391b745436fd9b2a19b3f74889a5965c24d27d1d818a0d49b66ac4f47a",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg1-1-3.src.rpm",
			Checksum:     "cf28031ec6d8ddc146b1f2ec37e4b7bbd7f4f751ae9b7730f709c43d0292bb79",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg1-1-3.x86_64.rpm",
			Checksum:     "d9c6b377f3484f5a6312164df290186d9ed5536a8e3e67ef1c7ae8c9b956794b",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg10-1-1.src.rpm",
			Checksum:     "7f86347249889174a4bcfa2d20aec471e44275be06484f45a2d1afa4f61d8895",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg10-1-1.x86_64.rpm",
			Checksum:     "af60a18cd0517920acac3ff737e5595c4b2dc779dd9205f6731eeba54265dfd7",
			ChecksumType: "sha256",
			Size:         6740,
		},
		{
			Name:         "SRPMS/testpkg10-1-2.src.rpm",
			Checksum:     "151a800766bfb51b8f414878367f64028f3693dc323a01de2ccc10d96813f302",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg10-1-2.x86_64.rpm",
			Checksum:     "6dc90f4258183bb61baa2a0a588732fa5db0fd165d1b7bf02ff0198fca93efdb",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg10-1-3.src.rpm",
			Checksum:     "370e830462c33db0cc11dcaa7f539773a651c1dde980bd1e92db601d85156419",
			ChecksumType: "sha256",
			Size:         6125,
		},
		{
			Name:         "RPMS/x86_64/testpkg10-1-3.x86_64.rpm",
			Checksum:     "a68d0510bec578428402a029eac55e34e8784e96556e9f2d1c9424911c2a489f",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg2-1-1.src.rpm",
			Checksum:     "802c968cd29794edf721fcbc5924a68606df9a0648d89c648481bf77ff89c53c",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg2-1-1.x86_64.rpm",
			Checksum:     "88307ad55751656bb96132d8d448aa78ebafb28120634695f8207c8460ad7dff",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg2-1-2.src.rpm",
			Checksum:     "f26e8390a958c0fd0a15c07cf9ef2f93485b90ec701888013f1453b0946c7c47",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg2-1-2.x86_64.rpm",
			Checksum:     "e5dd6aac17915a9ada1b39e4efcb03e4ee3e6998ff910c0584ccb11f97721632",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg2-1-3.src.rpm",
			Checksum:     "07a1e8b441c8e329cb07567bdbd845d833da6721dfc166c3510469f313ebd652",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg2-1-3.x86_64.rpm",
			Checksum:     "d216d06a4ff1569a98ff981469bbf2765300587f69dbf378a150c0cf8dfb4795",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg3-1-1.src.rpm",
			Checksum:     "3a557ebe499975f3c7b95e1b39ca3ff10369b51dec045d8f882a0cb587815db4",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg3-1-1.x86_64.rpm",
			Checksum:     "68170e526f756eeb06d33d77332000a67e8eb31eb004f28de509ed8aae72f8a6",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg3-1-2.src.rpm",
			Checksum:     "4ad8107d6a8a9f32c3c2f4971756467ee536979cb246e53a9d238c824df665ea",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg3-1-2.x86_64.rpm",
			Checksum:     "f75513d7fc853e3f8d4c832700151dcb04d88388f367c4a29a0c702988fd9c80",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg3-1-3.src.rpm",
			Checksum:     "14b797a378b75c5c6793400f2d2f14b98099ba32b95828686ca7b14c8475d889",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg3-1-3.x86_64.rpm",
			Checksum:     "ec036e47e7e711cf26b126c103e25bf0191a67bfc44089f18ee6c3ade8a334d1",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg4-1-1.src.rpm",
			Checksum:     "7b981767615dd5c739f027bcc17aa91343d36aa5040e59f7614e401a35b2e703",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg4-1-1.x86_64.rpm",
			Checksum:     "e7de03570c76ec13cceff7bca33cec6fcff45e3c4c83e088b06764825542e78b",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg4-1-2.src.rpm",
			Checksum:     "3f55e2dd76104d19baf4148342c0cdc5b0230879cca6c5b13c07f24a25d79a1d",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg4-1-2.x86_64.rpm",
			Checksum:     "c9731ea0936b1c4debc4ebae7935687c4a5a5f9b538ab795c903adbce6ba5b70",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg4-1-3.src.rpm",
			Checksum:     "80ea7ece7614339e906e80920a5a5d5bf0881f3e9a0a02ab2831036e67ea4152",
			ChecksumType: "sha256",
			Size:         6125,
		},
		{
			Name:         "RPMS/x86_64/testpkg4-1-3.x86_64.rpm",
			Checksum:     "5942505f4082335d8cde4e03bb9bb9080e86efb6a24dce6167a71abbe418a1d2",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg5-1-1.src.rpm",
			Checksum:     "182367519219c7f56c1213350372333219fa0d2f858072f524734308949daf61",
			ChecksumType: "sha256",
			Size:         6122,
		},
		{
			Name:         "RPMS/x86_64/testpkg5-1-1.x86_64.rpm",
			Checksum:     "006ef00d887654372886d295c1dc03eff7b18c611c1d35c4081456300bf99d15",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg5-1-2.src.rpm",
			Checksum:     "904021dcfd97e54d3c5cbaa7742af844cc5c48adb4af4ba88495db30afd629d9",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg5-1-2.x86_64.rpm",
			Checksum:     "ab195eaac0cc29c13c5920e0bdb8f2dedcc20ca10861775d6a0ccce8bb81546c",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg5-1-3.src.rpm",
			Checksum:     "7b74718b1f01b561f66d04706327269bf67981db6743c46858fe03e2868c6ef5",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg5-1-3.x86_64.rpm",
			Checksum:     "f2149561e84c54c429869ec051b18b3c59a92313164fc2b678c08b182711e339",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg6-1-1.src.rpm",
			Checksum:     "96e669c98d8d79d0acc77f1d9b3246e6baabf2f49a7591d26ba69740b80574ee",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg6-1-1.x86_64.rpm",
			Checksum:     "234b5d4a4ea3cd320292c374efdb2ca51876fac43f902d3816a8b04bd155a5c1",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg6-1-2.src.rpm",
			Checksum:     "798464bf773513c123183fe00aa350459ec31568e2ee1895d8fef4d6dfbb2e28",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg6-1-2.x86_64.rpm",
			Checksum:     "2665d47c785688562748b0d7c0637a2c9ff052cb12acf09a2b64d0da699843f2",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg6-1-3.src.rpm",
			Checksum:     "d9e1eaa79d2b80349186b3ec51a5331ca400994d9ba0a5fbcd6992fdf2234fcb",
			ChecksumType: "sha256",
			Size:         6125,
		},
		{
			Name:         "RPMS/x86_64/testpkg6-1-3.x86_64.rpm",
			Checksum:     "84c25528dc5381c89656e52a6af6fa51c7cb9faa69cec3b5c196ffd4ba037c80",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg7-1-1.src.rpm",
			Checksum:     "8f76905b96b5476485fca223804a9900c3351f0b61e30f862ebc7e48f1232fca",
			ChecksumType: "sha256",
			Size:         6123,
		},
		{
			Name:         "RPMS/x86_64/testpkg7-1-1.x86_64.rpm",
			Checksum:     "6b3edf4d9a1194b0b07c660281378375d258f25e1e74c7852f9fb209c0b1de4f",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg7-1-2.src.rpm",
			Checksum:     "263e504fd0be3d6903e03e4b5d3c820c6f06e474253d63bf7951223c3080cc17",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg7-1-2.x86_64.rpm",
			Checksum:     "8c9d92872e4c3c462203682c63c79d61c00b398f84a4adf66f2da08e4c416a4b",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg7-1-3.src.rpm",
			Checksum:     "c04e4e225556951502e2ce533ac25d26decf63a768013c14bcec32dc60b88837",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg7-1-3.x86_64.rpm",
			Checksum:     "845b31dbb4d80b13ace39ab2da1e19d68727d43de7b01fd5cf5c10e633f7707e",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg8-1-1.src.rpm",
			Checksum:     "6b0bb85fe30ae6e406a1dddb2af7126294e24fae64e79be67c426e94f90720ee",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg8-1-1.x86_64.rpm",
			Checksum:     "1ed2f5e08a1f57c68199c57271404945762bcba36d99c2c03a45d60ad8e53a75",
			ChecksumType: "sha256",
			Size:         6741,
		},
		{
			Name:         "SRPMS/testpkg8-1-2.src.rpm",
			Checksum:     "cc3fd1c43ea62332e534e0f5d6a7df7f1c808725116109f33fe38453682d5ee7",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg8-1-2.x86_64.rpm",
			Checksum:     "dde30ee7035800604434e3b0928934c56fc382748cb3dc6a87ee7605a15af9a3",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg8-1-3.src.rpm",
			Checksum:     "d5ccf0943d89947bb2f8b3ed8466d60e452f1b0daa881d5d3fe2c8584778883b",
			ChecksumType: "sha256",
			Size:         6125,
		},
		{
			Name:         "RPMS/x86_64/testpkg8-1-3.x86_64.rpm",
			Checksum:     "1fec4e8fcd788227fd836960f129b7c5384bf416aa78bdb7923dfdc431b208bf",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg9-1-1.src.rpm",
			Checksum:     "ed76da35a9b4f4ea4e3828f6a978016cc6b4b9c0cd0548f2d4cbd6a695b1f6e9",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg9-1-1.x86_64.rpm",
			Checksum:     "0ba29dded6b136adb548396295f46ff4fa827d416cdfa25129ddaef5be2f2d96",
			ChecksumType: "sha256",
			Size:         6740,
		},
		{
			Name:         "SRPMS/testpkg9-1-2.src.rpm",
			Checksum:     "26dc22dbbbd6569c6d41d7c37b6e8b3d90ec24c76c075bfe8b0016de64bb2029",
			ChecksumType: "sha256",
			Size:         6124,
		},
		{
			Name:         "RPMS/x86_64/testpkg9-1-2.x86_64.rpm",
			Checksum:     "192d106bc5c19b9e8651dbd255f5f0a240b781b85a2af0ce58d1c5f8670158b6",
			ChecksumType: "sha256",
			Size:         6742,
		},
		{
			Name:         "SRPMS/testpkg9-1-3.src.rpm",
			Checksum:     "869bf9b01e2ea53e02bb8666ca1118e2c9e5cef2598e9a1b1f67125531b3bd8e",
			ChecksumType: "sha256",
			Size:         6125,
		},
		{
			Name:         "RPMS/x86_64/testpkg9-1-3.x86_64.rpm",
			Checksum:     "ae2f90464235b2abeccf9c2607c51bab0b8a4f3afd6531026d62d54300b2e093",
			ChecksumType: "sha256",
			Size:         6742,
		},
	}, packages)
}

func TestMetadataSHA256(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	_, err := repo.Packages(context.Background())
	r.NoError(err)

	type mdFile struct {
		size     int
		checksum string
	}

	md := repo.Metadata()
	r.Equal(map[string]mdFile{
		"repodata/3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5-primary.sqlite.bz2": {
			size: 9036, checksum: "3228a79e4328e912dc272ea29a6e627187e4d1ecdb41c39f17e67b0a4f1ea1a5",
		},
		"repodata/12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053-primary.xml.gz": {
			size: 4550, checksum: "12fd2c7242e8f946c5a99b40acc94e297144685022f23f08f5d4932a37387053",
		},
		"repodata/a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52-filelists.sqlite.bz2": {
			size: 5554, checksum: "a362c47ae86d8ee5f6e623fcb09e309d5b701316fe356e7642fa24ca47610c52",
		},
		"repodata/f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd-filelists.xml.gz": {
			size: 3000, checksum: "f47c809e5e0589800a672ee64fd4f4fd7bc71152596bcf8ebcd6a582d40ca9dd",
		},
		"repodata/27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32-other.sqlite.bz2": {
			size: 4351, checksum: "27c7a0b93802b0173730c713cf74c0ac80e402880e3a87f12c5dc3efcfa0ec32",
		},
		"repodata/3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f-other.xml.gz": {
			size: 2848, checksum: "3af4f502859773f9b5b5a3b8033f8561f01b270414cabb3b46612703c85d731f",
		},
		"repodata/repomd.xml": {
			size: 3078, checksum: "904c00f4c838f67d1c79113d7996840add665d513889b112bb715776607c151c",
		},
	}, func() map[string]mdFile {
		keys := map[string]mdFile{}
		for k, v := range md {
			h := sha256.New()
			n, err := h.Write(v)
			if err != nil {
				panic(err)
			}

			if n != len(v) {
				panic(io.ErrShortWrite)
			}

			keys[k] = mdFile{
				size:     len(v),
				checksum: hex.EncodeToString(h.Sum(nil)),
			}
		}
		return keys
	}())
}

func TestMetadataSHA1(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo-sha1")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	_, err := repo.Packages(context.Background())
	r.NoError(err)

	md := repo.Metadata()
	r.Equal(map[string]int{
		"repodata/repomd.xml": 2601,
		"repodata/4a11e3eeb25d21b08f41e5578d702d2bea21a2e7-filelists.xml.gz":     282,
		"repodata/fdedb6ce109127d52228d01b0239010ddca14c8f-other.xml.gz":         247,
		"repodata/e7a8a53e7398f6c22894718ea227fea60f2b78ba-primary.sqlite.bz2":   1937,
		"repodata/c66ce2caa41ed83879f9b3dd9f40e61c65af499e-filelists.sqlite.bz2": 787,
		"repodata/b31561a27d014d35b59b27c27859bb1c17ac573e-other.sqlite.bz2":     669,
		"repodata/80779e2ab55e25a77124d370de1d08deae8f1cc6-primary.xml.gz":       688,
	}, func() map[string]int {
		keys := map[string]int{}
		for k, v := range md {
			keys[k] = len(v)
		}
		return keys
	}())
}
