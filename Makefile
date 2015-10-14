SHA := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat VERSION)
ITTERATION := $(shell date +%s)

build:
	mkdir -p packaging/output
	mkdir -p packaging/root/usr/local/bin
	go build -o packaging/root/usr/local/bin/synd github.com/pandemicsyn/syndicate/synd

clean:
	rm -rf packaging/output
	rm -f packaging/root/usr/local/bin/synd

install:
	install -t /usr/local/bin packaging/root/usr/local/bin/synd

run:
	go run synd/*.go

ring:
	mkdir -p /etc/oort/ring
	ring /etc/oort/ring/oort.builder create replicas=3
	ring /etc/oort/ring/oort.builder add active=true capacity=1000 tier0=server3 address0=127.0.0.1:8003 address1=127.0.0.2:8003 meta=onmetalv1
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
