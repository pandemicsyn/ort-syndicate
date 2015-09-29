SHA := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat VERSION)
ITTERATION := $(shell date +%s)

build:
	mkdir -p packaging/output
	mkdir -p packaging/root/usr/local/bin
	go build -o packaging/root/usr/local/bin/synd github.com/pandemicsyn/ort-syndicate/synd

clean:
	rm -rf packaging/output
	rm -f packaging/root/usr/local/bin/synd

install:
	install -t /usr/local/bin packaging/root/usr/local/bin/synd

run:
	go run synd/*.go

ring:
	ring /etc/ort/ring/ort.builder create replicas=3
	ring /etc/ort/ring/ort.builder add active=true capacity=1000 tier0=server3 address0=127.0.0.1:8003 address1=127.0.0.2:8003 meta=onmetalv1
	ring /etc/ort/ring/ort.builder ring

packages: clean build deb

deb:
	fpm -s dir -t deb -n ort-syndicate -v $(VERSION) -p packaging/output/ort-syndicate-$(VERSION)_amd64.deb \
		--deb-priority optional --category admin \
		--force \
		--iteration $(ITTERATION) \
		--deb-compression bzip2 \
	 	--after-install packaging/scripts/postinst.deb \
	 	--before-remove packaging/scripts/prerm.deb \
		--after-remove packaging/scripts/postrm.deb \
		--url https://github.com/pandemicsyn/ort-syndicate \
		--description "Ring/config controller for ort" \
		-m "Florian Hines <syn@ronin.io>" \
		--license "Apache License 2.0" \
		--vendor "ort" -a amd64 \
		--config-files /etc/ort/syndicate.toml-sample \
		packaging/root/=/
	cp packaging/output/ort-syndicate-$(VERSION)_amd64.deb packaging/output/ort-syndicate.deb.$(SHA)
