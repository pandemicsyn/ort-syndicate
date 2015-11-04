#!/bin/bash

GOVERSION=1.5.1

echo "Using $GIT_USER as user"

echo "Setting up dev env"

apt-get update
apt-get install -y --force-yes vim git build-essential autoconf libtool
update-alternative -set editor /usr/bin/vim.basic

# setup go
mkdir -p /$USER/go/bin
export GVERSION=1.5.1
cd /tmp &&  wget -q https://storage.googleapis.com/golang/go$GVERSION.linux-amd64.tar.gz
tar -C /usr/local -xzf /tmp/go$GVERSION.linux-amd64.tar.gz
echo " " >> /$USER/.bashrc
echo "# Go stuff" >> /$USER/.bashrc
echo "export PATH=\$PATH:/usr/local/go/bin" >> /$USER/.bashrc
echo "export GOPATH=/root/go" >> /$USER/.bashrc
echo "export PATH=\$PATH:\$GOPATH/bin" >> /$USER/.bashrc
source /$USER/.bashrc

# setup protobuf
git clone https://github.com/google/protobuf.git
cd protobuf
./autogen.sh
./configure && make && make check && make install
ldconfig

# setup grpc
echo “deb http://http.debian.net/debian jessie-backports main” >> /etc/apt/sources.list
apt-get update
apt-get install libgrpc-dev
go get google.golang.org/grpc
go get github.com/golang/protobuf/proto
go get github.com/golang/protobuf/protoc-gen-go

# get syndicate and setup repos
mkdir -p $GOPATH/src/github.com/pandemicsyn
cd $GOPATH/src/github.com/pandemicsyn/
git clone git@github.com:$GIT_USER/syndicate.git
cd syndicate
git remote add upstream git@github.com:pandemicsyn/syndicate.git

# get oort and setup repos
cd $GOPATH/src/github.com/pandemicsyn
git clone git@github.com:$GIT_USER/oort.git
cd oort
git remote add upstream git@github.com:pandemicsyn/oort.git


# install syndicate and ort
cd $GOPATH/src/github.com/pandemicsyn/syndicate
make allinone
cp -av packaging/root/usr/share/oort/systemd/synd.service /lib/systemd/system 
systemctl daemon-reload

go get github.com/pandemicsyn/oort/oortd
go install github.com/pandemicsyn/oort/oortd
cp -av packaging/root/usr/share/oort/systemd/oortd.service /lib/systemd/system 
systemctl daemon-reload

# setup apid deps
go get github.com/pandemicsyn/oort/apid
# setup cfs deps
go get github.com/pandemicsyn/oort/cfs
