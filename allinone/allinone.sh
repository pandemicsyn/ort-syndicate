#!/bin/bash

GOVERSION=1.5.1

echo "Using $GIT_USER as user"

echo "Setting up dev env"

apt-get update
apt-get install -y --force-yes vim git build-essential autoconf libtool libtool-bin unzip fuse
update-alternatives --set editor /usr/bin/vim.basic

# setup grpc
echo deb http://http.debian.net/debian jessie-backports main >> /etc/apt/sources.list
apt-get update
apt-get install libgrpc-dev -y --force-yes

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

# sup3r sekret option to install vim-go and basic vim-go friendly .vimrc
if [ "$FANCYVIM" = "yes" ]; then
    mkdir -p ~/.vim/autoload ~/.vim/bundle && curl -LSso ~/.vim/autoload/pathogen.vim https://tpo.pe/pathogen.vim
    git clone https://github.com/fatih/vim-go.git ~/.vim/bundle/vim-go
    apt-get install -y --force-yes vim-youcompleteme
    go get golang.org/x/tools/cmd/goimports
    git clone https://github.com/fatih/vim-go.git ~/.vim/bundle/vim-go
    curl -o ~/.vimrc https://raw.githubusercontent.com/pandemicsyn/syndicate/master/allinone/.vimrc 
    go get code.google.com/p/rog-go/exp/cmd/godef
    go get github.com/nsf/gocode
else
    echo "You didn't set FANCYVIM=yes so no awesome vim-go setup for you."
fi

# setup protobuf
if [ "$BUILDPROTOBUF" = "yes" ]; then
    echo "Building with protobuf support, this gonna take awhile"
    cd $HOME
    git clone https://github.com/google/protobuf.git
    cd protobuf
    ./autogen.sh && ./configure && make && make check && make install && ldconfig
else 
    echo "Built withOUT protobuf"
fi

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

# get formic/cfs and setup repos
mkdir -o $GOPATH/src/github.com/creiht
cd $GOPATH/src/github.com/creiht
git clone git@github.com:$GIT_USER/formic.git
cd formic
git remote add upstream git@github.com:creiht/formic.git


# install syndicate and ort
cd $GOPATH/src/github.com/pandemicsyn/syndicate
mkdir -p /etc/oort/ring
mkdir -p /etc/oort/oortd
cp -av allinone/etc/oort/* /etc/oort
make deps
go install github.com/gholt/ring/ring
ring /etc/oort/ring/oort.builder create replicas=1 configfile=/etc/oort/oort.toml
ring /etc/oort/ring/oort.builder add active=true capacity=1000 tier0=removeme
ring /etc/oort/ring/oort.builder ring
go get github.com/pandemicsyn/ringver
go install github.com/pandemicsyn/ringver
RINGVER=`ringver /etc/oort/ring/oort.ring`
cp -av /etc/oort/ring/oort.ring /etc/oort/ring/$RINGVER-oort.ring
cp -av /etc/oort/ring/oort.builder /etc/oort/ring/$RINGVER-oort.builder
cp -av packaging/root/usr/share/syndicate/systemd/synd.service /lib/systemd/system 
go get github.com/pandemicsyn/syndicate/synd
make install
systemctl daemon-reload

go get github.com/pandemicsyn/oort/oortd
go install github.com/pandemicsyn/oort/oortd
cd $GOPATH/src/github.com/pandemicsyn/oort
cp -av packaging/root/usr/share/oort/systemd/oortd.service /lib/systemd/system 
echo "OORT_SYNDICATE_OVERRIDE=127.0.0.1:8443" >> /etc/default/oortd
systemctl daemon-reload

# setup formic deps
go get github.com/creiht/formic/formic
# setup cfs deps
go get github.com/creiht/formic/cfs

echo "To start services run:"
echo "systemctl start synd"
echo "systemctl start oortd"

