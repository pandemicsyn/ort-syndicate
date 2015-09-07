# ort-syndicate

### basic installation

```
go get -u github.com/pandemicsyn/ort-syndicate/synd
go install -a github.com/pandemicsyn/ort-syndicate/synd
mkdir -p /etc/ort
cp -av ~/go/src/github.com/pandemicsyn/ort-syndicate/synd/server.crt /etc/ort
cp -av ~/go/src/github.com/pandemicsyn/ort-syndicate/synd/server.key /etc/ort
```

### setup

You need to make sure:

- /etc/ort exists
- /etc/ort/ring exists
- /etc/ort/ring/ort.builder exists and has at least one active node
- /etc/ort/ring/ort.ring exists and has at least one active node
- /etc/ort contains valid server.crt and server.key

### temporary dev step (this will go away)

The first time you try and start synd you'll get an error like:

```
root@syndicate1:~/go/src/github.com/pandemicsyn/ort-syndicate/synd# go run *.go -master=true 
2015/09/07 19:56:18 open /etc/ort/syndicate.toml: no such file or directory
2015/09/07 19:56:18 Using default net filter: [10.0.0.0/8 192.168.0.0/16]
2015/09/07 19:56:18 Using default tier filter: [z.*]
2015/09/07 19:56:18 Found /etc/ort/ring/ort.builder, as last builder
2015/09/07 19:56:18 Found /etc/ort/ring/ort.ring, as last ring
2015/09/07 19:56:18 Ring version is: 1439924753845934663
2015/09/07 19:56:18 Attempting to load ring/builder bytes: open /etc/ort/ring/1439924753845934663-ort.builder: no such file or directory
exit status 1
```

To fix this: 

- `cp -av /etc/ring/ort.builder /etc/ort/ring/$THERINGVERSION-ort.builder`
- `cp -av /etc/ring/ort.ring /etc/ort/ring/$THERINGVERSION-ort.ring`

If it works you'll see something along the lines of:

```
fhines@47:~/go/src/github.com/pandemicsyn/ort-syndicate/synd (master)$ go build . && ./synd -master=true -ring_dir=/etc/ort/ring
2015/09/07 19:58:38 open /etc/ort/syndicate.toml: no such file or directory
2015/09/07 19:58:38 Using default net filter: [10.0.0.0/8 192.168.0.0/16]
2015/09/07 19:58:38 Using default tier filter: [z.*]
2015/09/07 19:58:38 Found /etc/ort/ring/1439924753845934663-ort.builder, as last builder
2015/09/07 19:58:38 Found /etc/ort/ring/1439924753845934663-ort.ring, as last ring
2015/09/07 19:58:38 Ring version is: 1439924753845934663
2015/09/07 19:58:38 !! Running without slaves, have no one to register !!
2015/09/07 19:58:38 Master starting up on 8443...
```

### slaves

aren't working yet

### systemd init script

A working systemd init script is provided in packaging/root/usr/share/synd/systemd/synd.service. To use it
on Debian Jessie `cp packaging/root/usr/share/ort-syndicate/systemd/synd.service /lib/systemd/system`. You can then
stop/start/restart the synd service as usual ala `systemctl start synd`. its setup to capture and log to syslog.

### building packages

should be possible if you have fpm install. try `make packages`

