# syndicate

### setup

You need to make sure:

- /etc/oort exists
- /etc/oort/ring exists
- /etc/oort/ring/oort.builder exists and has at least one active node
- /etc/oort/ring/oort.ring exists and has at least one active node
- /etc/oort contains valid server.crt and server.key
- /etc/oort/syndicate.toml contains a valid config like:
```
master = "true"
ringdir = "/etc/oort/ring"
certfile = "/etc/oort/server.crt"
keyfile = "/etc/oort/server.key"
tls = "true"
```
- /etc/oort/ring/oort.toml contains a valid config like:
```
[CmdCtrlConfig]
ListenAddress = "0.0.0.0:4444"
CertFile = "/etc/oort/server.crt"
KeyFile = "/etc/oort/server.key"
UseTLS = true
Enabled = true

[ValueStoreConfig]
CompactionInterval = 42
ValueCap = 4194302

[TCPMsgRingConfig]
UseTLS = true
CertFile = "/etc/oort/server.crt"
KeyFile = "/etc/oort/server.key"
```

### temporary dev step (this will go away)

The first time you try and start synd you'll get an error like:

```
root@syndicate1:~/go/src/github.com/pandemicsyn/syndicate/synd# go run *.go -master=true 
2015/09/07 19:56:18 open /etc/oort/syndicate.toml: no such file or directory
2015/09/07 19:56:18 Using default net filter: [10.0.0.0/8 192.168.0.0/16]
2015/09/07 19:56:18 Using default tier filter: [z.*]
2015/09/07 19:56:18 Found /etc/oort/ring/oort.builder, as last builder
2015/09/07 19:56:18 Found /etc/oort/ring/oort.ring, as last ring
2015/09/07 19:56:18 Ring version is: 1439924753845934663
2015/09/07 19:56:18 Attempting to load ring/builder bytes: open /etc/oort/ring/1439924753845934663-oort.builder: no such file or directory
exit status 1
```

To fix this: 

- `cp -av /etc/ring/oort.builder /etc/oort/ring/$THERINGVERSION-oort.builder`
- `cp -av /etc/ring/oort.ring /etc/oort/ring/$THERINGVERSION-oort.ring`

If it works you'll see something along the lines of:

```
fhines@47:~/go/src/github.com/pandemicsyn/syndicate/synd (master)$ go build . && ./synd -master=true -ring_dir=/etc/oort/ring
2015/09/07 19:58:38 open /etc/oort/syndicate.toml: no such file or directory
2015/09/07 19:58:38 Using default net filter: [10.0.0.0/8 192.168.0.0/16]
2015/09/07 19:58:38 Using default tier filter: [z.*]
2015/09/07 19:58:38 Found /etc/oort/ring/1439924753845934663-oort.builder, as last builder
2015/09/07 19:58:38 Found /etc/oort/ring/1439924753845934663-oort.ring, as last ring
2015/09/07 19:58:38 Ring version is: 1439924753845934663
2015/09/07 19:58:38 !! Running without slaves, have no one to register !!
2015/09/07 19:58:38 Master starting up on 8443...
```

### syndicate-client

go install github.com/pandemicsyn/syndicate/syndicate-client


```
    syndicate-client
        Valid commands are:
        version         #print version
        config          #print ring config
        config <nodeid> #uses uint64 id
        search          #lists all
        search id=<nodeid>
        search meta=<metastring>
        search tier=<string> or search tierX=<string>
        search address=<string> or search addressX=<string>
        search any of the above K/V combos
        rm <nodeid>
        set config=./path/to/config
```

### Oortd 

Oort either needs a valid SVR record setup or you need to set OORT_SYNDICATE_OVERRIDE=127.0.0.1:8443 when running oortd.

### slaves

aren't working yet

### systemd init script

A working systemd init script is provided in packaging/root/usr/share/synd/systemd/synd.service. To use it
on Debian Jessie `cp packaging/root/usr/share/syndicate/systemd/synd.service /lib/systemd/system`. You can then
stop/start/restart the synd service as usual ala `systemctl start synd`. its setup to capture and log to syslog.

### building packages

should be possible if you have fpm install. try `make packages`

### drone config
