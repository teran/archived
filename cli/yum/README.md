# cli/service

## Tests notice

Tests for cli/service are using samples with complete yum repositories so
require some data generation for yum package.

### Test environment

#### Base image

```shell
docker run -it rockylinux/rockylinux:9.4
```

#### Dependencies

```shell
dnf install -y \
    createrepo_c \
    pinentry \
    rpm-build \
    rpm-sign \
    vim
```

#### Generate GPG key

```shell
gpg --gen-key
```

Follow instructions but please remember to leave something meaning in key name
like the following name: `archived test repository sample`

Export GPG public key for further tests update:

```shell
gpg --export -a 'archived test repository sample' > RPM-GPG-KEY-testrepo
```

#### Fill .rpmmacros

~/.rpmmacros:

```text
%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name archived test repository sample
%_gpgbin /usr/bin/gpg2
```

#### Generate packages

testpkg.spec:

```shell
Name:       testpkg
Version:    1
Release:    1
Summary:    Test RPM package
License:    FIXME

%description
Test RPM package

%prep

%build
cat > hello-world.sh <<EOF
#!/usr/bin/bash

set -euo pipefail

echo "Hello world"
EOF

%install
mkdir -p %{buildroot}/usr/bin/
install -m 755 hello-world.sh %{buildroot}/usr/bin/hello-world.sh

%files
/usr/bin/hello-world.sh

%changelog
```

```shell
rpmbuild -ba testpkg.spec
```

#### Sign package

```shell
rpm --addsign rpmbuild/RPMS/x86_64/testpkg-1-1.x86_64.rpm
rpm --addsign rpmbuild/SRPMS/testpkg-1-1.src.rpm
```

#### Generate repository metadata

```shell
mkdir repo
cp -vr {rpmbuild/RPMS,rpmbuild/SRPMS} repo/
cd repo && createrepo .
```

All done but note this almost definitely will require to update tests!
