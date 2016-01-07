#!/bin/bash

GOVERSION=1.5.1

echo "Using $GIT_USER as user"

echo "Setting up dev env"

apt-get update
apt-get install -y --force-yes vim git build-essential autoconf libtool libtool-bin unzip fuse mercurial
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
    echo "Performing fancy vim install"
    mkdir -p ~/.vim/autoload ~/.vim/bundle && curl -LSso ~/.vim/autoload/pathogen.vim https://tpo.pe/pathogen.vim
    git clone https://github.com/fatih/vim-go.git ~/.vim/bundle/vim-go
    apt-get install -y --force-yes vim-youcompleteme
    go get golang.org/x/tools/cmd/goimports
    git clone https://github.com/fatih/vim-go.git ~/.vim/bundle/vim-go
    curl -o ~/.vimrc https://raw.githubusercontent.com/pandemicsyn/syndicate/master/allinone/.vimrc 
    go get github.com/nsf/gocode
    echo "Fancy VIM install complete. You may way want to open vim and run ':GoInstallBinaries' the first time you use it"
    sleep 1
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
    echo "Protobuf build done...hopefully"
else 
    echo "Built withOUT protobuf"
fi

echo "Setting up the imporant bits..."
go get google.golang.org/grpc
go get github.com/golang/protobuf/proto
go get github.com/golang/protobuf/protoc-gen-go

echo "Setting up syndicate repo"
mkdir -p $GOPATH/src/github.com/pandemicsyn
cd $GOPATH/src/github.com/pandemicsyn/
git clone git@github.com:$GIT_USER/syndicate.git
cd syndicate
git remote add upstream git@github.com:pandemicsyn/syndicate.git

echo "Setting up oort repos"
cd $GOPATH/src/github.com/pandemicsyn
git clone git@github.com:$GIT_USER/oort.git
cd oort
git remote add upstream git@github.com:pandemicsyn/oort.git

echo "Setting up formic/cfs repos"
mkdir -p $GOPATH/src/github.com/creiht
cd $GOPATH/src/github.com/creiht
git clone git@github.com:$GIT_USER/formic.git
cd formic
git remote add upstream git@github.com:creiht/formic.git


echo "Prepping /etc"
cd $GOPATH/src/github.com/pandemicsyn/syndicate
mkdir -p /etc/oort/ring
mkdir -p /etc/oort/value /etc/oort/group
cp -av allinone/etc/oort/* /etc/oort

echo "Install go deps"
make deps
go install github.com/gholt/ring/ring
go get github.com/pandemicsyn/ringver
go install github.com/pandemicsyn/ringver

echo "Setting up valuestore rings"
mkdir -p /etc/oort/ring/value
ring /etc/oort/ring/value/valuestore.builder create replicas=1 configfile=/etc/oort/valuestore.toml
ring /etc/oort/ring/value/valuestore.builder add active=true capacity=1000 tier0=removeme
ring /etc/oort/ring/value/valuestore.builder ring
RINGVER=`ringver /etc/oort/ring/value/valuestore.ring`
cp -av /etc/oort/ring/value/valuestore.ring /etc/oort/ring/value/$RINGVER-valuestore.ring
cp -av /etc/oort/ring/value/valuestore.builder /etc/oort/ring/value/$RINGVER-valuestore.builder

echo "Setting up groupstore rings"
mkdir -p /etc/oort/ring/group
ring /etc/oort/ring/group/groupstore.builder create replicas=1 configfile=/etc/oort/groupstore.toml
ring /etc/oort/ring/group/groupstore.builder add active=true capacity=1000 tier0=removeme
ring /etc/oort/ring/group/groupstore.builder ring
RINGVER=`ringver /etc/oort/ring/group/groupstore.ring`
cp -av /etc/oort/ring/group/groupstore.ring /etc/oort/ring/group/$RINGVER-groupstore.ring
cp -av /etc/oort/ring/group/groupstore.builder /etc/oort/ring/group/$RINGVER-groupstore.builder

echo "Installing synd"
cp -av packaging/root/usr/share/syndicate/systemd/synd.service /lib/systemd/system 
go get github.com/pandemicsyn/syndicate/synd
make install
systemctl daemon-reload

echo "Installing oort-valued"
go get github.com/pandemicsyn/oort/oort-valued
go install github.com/pandemicsyn/oort/oort-valued
cd $GOPATH/src/github.com/pandemicsyn/oort
cp -av packaging/root/usr/share/oort/systemd/oort-valued.service /lib/systemd/system 
echo "OORT_VALUE_SYNDICATE_OVERRIDE=127.0.0.1:8443" >> /etc/default/oort-valued
systemctl daemon-reload

echo "Installing oort-groupd"
go get github.com/pandemicsyn/oort/oort-groupd
go install github.com/pandemicsyn/oort/oort-groupd
cd $GOPATH/src/github.com/pandemicsyn/oort
cp -av packaging/root/usr/share/oort/systemd/oort-groupd.service /lib/systemd/system 
echo "OORT_GROUP_SYNDICATE_OVERRIDE=127.0.0.1:8443" >> /etc/default/oort-groupd
systemctl daemon-reload

echo "Creating data dir"
mkdir -v -p /data

# setup formic & cfs deps
go get github.com/creiht/formic/formicd
go install github.com/creiht/formic/formicd
go get github.com/creiht/formic/cfs
go install github.com/creiht/formic/cfs
cp -av go/src/github.com/creiht/formic/packaging/root/usr/share/formicd/systemd/formicd.service /lib/systemd/system
echo 'FORMICD_PORT=8444' > /etc/default/formicd

# Adding some helpful git stuff to the .bashrc 
echo "" >> ~/.bashrc
echo "# Added to show git branches" >> ~/.bashrc
echo 'export PS1="\u@\h \W\[\033[37m\]\$(git_branch)\[\033[00m\] $ "' >> ~/.bashrc
echo '' >> ~/.bashrc
echo '# get the current git branch' >> ~/.bashrc
echo 'git_branch() {' >> ~/.bashrc
echo "        git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/ (\1)/'" >> ~/.bashrc
echo '    }' >> ~/.bashrc



echo 
echo "To start services run:"
echo "systemctl start synd"
echo "systemctl start oort-valued"
echo "systemctl start formicd"
echo ""
echo "For example: to create a cfsfuse mount point create the location and run cfs:"
echo "mkdir -p /mnt/cfs"
echo "cfs -host localhost:8445 /mnt/cfs"
echo ""
echo "If you plan on using *THIS* session and to get the git enhanced prompt make sure to source ~/.bashrc to load path changes"

