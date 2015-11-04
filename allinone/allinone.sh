#!/bin/bash

GOVERSION=1.5.1

echo "Using $GIT_USER as user"

echo "Setting up dev env"

apt-get update
apt-get install -yes vim
update-alternative -set editor /usr/bin/vim.basic
apt-get install git build-essential autoconf libtool

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
cd protobuf.git
./autogen.sh
./configure && make && make check && make install
ldconfig

# setup grpc
DISTRIB=`lsb_release -c -s`
if [ "$DISTRIBG" == "jessie" ]; then
    echo “deb http://http.debian.net/debian jessie-backports main” >> /etc/apt/sources.list
    apt-get update
    apt-get install libgrpc-dev
else 
    echo "!! Y U NO USE JESSIE? Now you have to build grpc yourself !!"
fi
go get google.golang.org/grpc
go get github.com/golang/protobuf/proto
go get github.com/golang/protobuf/protoc-gen-go

# setup syndicate
mkdir -p $GOPATH/src/github.com/pandemicsyn
cd $GOPATH/src/github.com/pandemicsyn/
git clone git@github.com:$GIT_USER/syndicate.git
cd syndicate
git remote add upstream https://github.com/pandemicsyn/syndicate.git
make allinone
cp -av packaging/root/usr/share/oort/systemd/synd.service /lib/systemd/system 
systemctl daemon-reload

# setup oort
cd $GOPATH/src/github.com/pandemicsyn
git clone git@github.com:$GIT_USER/oort.git
cd oort
git remote add upstream git@github.com:pandemicsyn/oort.git
go get github.com/pandemicsyn/oort/oortd
go install github.com/pandemicsyn/oort/oortd
cp -av packaging/root/usr/share/oort/systemd/oortd.service /lib/systemd/system 
systemctl daemon-reload

# setup apid deps
go get github.com/pandemicsyn/oort/apid
# setup cfs deps
go get github.com/pandemicsyn/oort/cfs
