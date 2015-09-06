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
- /etc/ort.builder exists and has at least one active node
- /etc/ort.ring exists and has at least one active node
- /etc/ort contains valid server.crt and server.key

### temporary dev step (this will go away)

The first time you try and start synd you'll get an error like:

```
root@syndicate1:~/go/src/github.com/pandemicsyn/ort-syndicate/synd# go run *.go -master=true
2015/09/06 19:44:51 using /etc/ort/ort.builder
2015/09/06 19:44:51 Ring version is: 1439924753845934663
2015/09/06 19:44:51 Attempting to load ring/builder bytes: open /etc/ort/1439924753845934663-ort.builder: no such file or directory
exit status 1
```

To fix this: 

- `cp -av /etc/ort.builder /etc/ort/$THERINGVERSION-ort.builder`
- `cp -av /etc/ort.ring /etc/ort/$THERINGVERSION-ort.ring`

### slaves

aren't working yet

### systemd init script

A working systemd init script is provided in packaging/root/usr/share/synd/systemd/synd.service. To use it
on Debian Jessie `cp packaging/root/usr/share/synd/systemd/synd.service /lib/systemd/system`. You can then
stop/start/restart the synd service as usual ala `systemctl start synd`. its setup to capture and log to syslog.

