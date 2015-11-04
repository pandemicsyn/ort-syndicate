SHA := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat VERSION)
ITTERATION := $(shell date +%s)

allinone: deps localbuild install

localbuild:
	go get github.com/pandemicsyn/synd
	go install github.com/gholt/ring/ring
	mkdir -p /etc/oort/ring
	mkdir -p /etc/oort/oortd
	cp -av allinone/etc/oort/* /etc/oort
	go install github.com/gholt/ring/ring
	ring /etc/oort/ring/oort.builder create replicas=1 configfile=/etc/oort/oort.toml
	ring /etc/oort/ring/oort.builder add active=true capacity=1000 tier0=removeme 
	ring /etc/oort/ring/oort.builder ring
	go get github.com/pandemicsyn/ringver
	go install github.com/pandemicsyn/ringver
	RINGVER = $(shell ringver /etc/oort/ring/oort.ring)
	cp -av /etc/oort/ring/oort.ring /etc/oort/ring/$(RINGVER)-oort.ring
	cp -av /etc/oort/ring/oort.builder /etc/oort/ring/$(RINGVER)-oort.builder

deps:
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get github.com/gholt/ring
	go get github.com/gholt/ring/ring
	go get github.com/gholt/store

build:
	mkdir -p packaging/output
	mkdir -p packaging/root/usr/local/bin
	go build -o packaging/root/usr/local/bin/synd github.com/pandemicsyn/syndicate/synd
	go build -o packaging/root/usr/local/bin/syndicate-client github.com/pandemicsyn/syndicate/syndicate-client

clean:
	rm -rf packaging/output
	rm -f packaging/root/usr/local/bin/synd
	rm -f packaging/root/usr/local/bin/syndicate-client

install:
	#install -t /usr/local/bin packaging/root/usr/local/bin/synd
	go install github.com/pandemicsyn/syndicate/synd
	go install github.com/pandemicsyn/syndicate/syndicate-client

run:
	go run synd/*.go

ring:
	go get github.com/gholt/ring/ring
	go install github.com/gholt/ring/ring
	mkdir -p /etc/oort/ring
	ring /etc/oort/ring/oort.builder create replicas=3
	ring /etc/oort/ring/oort.builder add active=true capacity=1000 tier0=removeme 
	ring /etc/oort/ring/oort.builder ring

packages: clean build deb

deb:
	fpm -s dir -t deb -n syndicate -v $(VERSION) -p packaging/output/syndicate-$(VERSION)_amd64.deb \
		--deb-priority optional --category admin \
		--force \
		--iteration $(ITTERATION) \
		--deb-compression bzip2 \
	 	--after-install packaging/scripts/postinst.deb \
	 	--before-remove packaging/scripts/prerm.deb \
		--after-remove packaging/scripts/postrm.deb \
		--url https://github.com/pandemicsyn/syndicate \
		--description "Ring/config controller for Oort" \
		-m "Florian Hines <syn@ronin.io>" \
		--license "Apache License 2.0" \
		--vendor "Oort" -a amd64 \
		--config-files /etc/oort/syndicate.toml-sample \
		packaging/root/=/
	cp packaging/output/syndicate-$(VERSION)_amd64.deb packaging/output/syndicate.deb.$(SHA)
