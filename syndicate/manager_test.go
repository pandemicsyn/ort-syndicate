package syndicate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/gholt/ring"
	pb "github.com/pandemicsyn/syndicate/api/proto"
	"golang.org/x/net/context"
)

type MockRingBuilderThings struct {
	builderPath       string
	ringPath          string
	persistBuilderErr error
	persistRingErr    error
	bytesLoaderData   []byte
	bytesLoaderErr    error
	ring              ring.Ring
	ringbytes         *[]byte
	builder           *ring.Builder
	builderbytes      *[]byte
	buildererr        error
	managedNodes      map[uint64]ManagedNode
	slaves            []*RingSlave
	changeChan        chan *changeMsg
}

func (f *MockRingBuilderThings) BytesLoader(path string) ([]byte, error) {
	return f.bytesLoaderData, f.bytesLoaderErr
}

func (f *MockRingBuilderThings) Persist(c *RingChange, renameMaster bool) (error, error) {
	log.Println("Persist called with", c, renameMaster)
	return f.persistBuilderErr, f.persistRingErr
}

func (f *MockRingBuilderThings) GetBuilder(path string) (*ring.Builder, error) {
	return f.builder, f.buildererr
}

func newTestServerWithDefaults() (*Server, *MockRingBuilderThings) {
	b := ring.NewBuilder(64)
	b.SetReplicaCount(3)
	b.AddNode(true, 1, []string{"server1", "zone1"}, []string{"1.2.3.4:56789"}, "server1|meta one", []byte("Conf"))
	b.AddNode(true, 1, []string{"dummy1", "zone42"}, []string{"1.42.42.42:56789"}, "dummy1|meta one", []byte("Conf"))
	ring := b.Ring()

	rbytes := []byte("imnotaring")
	bbytes := []byte("imnotbuilder")

	mock := &MockRingBuilderThings{
		builderPath:  "/tmp/test.builder",
		ringPath:     "/tmp/test.ring",
		ring:         ring,
		ringbytes:    &rbytes,
		builder:      b,
		builderbytes: &bbytes,
		managedNodes: make(map[uint64]ManagedNode, 0),
		slaves:       make([]*RingSlave, 0),
		changeChan:   make(chan *changeMsg, 1),
	}
	s := newTestServer(&Config{}, "test", mock)
	_, netblock, _ := net.ParseCIDR("10.0.0.0/24")
	s.netlimits = append(s.netlimits, netblock)
	_, netblock, _ = net.ParseCIDR("1.2.3.0/24")
	s.netlimits = append(s.netlimits, netblock)
	return s, mock
}

func newTestServer(cfg *Config, servicename string, mockinfo *MockRingBuilderThings) *Server {
	s := &Server{}
	s.cfg = cfg
	s.servicename = servicename
	//s.parseConfig()

	s.rbPersistFn = mockinfo.Persist
	s.rbLoaderFn = mockinfo.BytesLoader
	s.getBuilderFn = mockinfo.GetBuilder
	s.b = mockinfo.builder
	s.r = mockinfo.ring
	s.rb = mockinfo.ringbytes
	s.bb = mockinfo.builderbytes
	s.managedNodes = make(map[uint64]ManagedNode, 0)
	s.slaves = mockinfo.slaves
	s.changeChan = mockinfo.changeChan
	return s
}

func TestWithRingBuilderPersister(t *testing.T) {
}

func TestWithRingBuilderBytesLoader(t *testing.T) {
}

func TestNewServer(t *testing.T) {
}

func TestServer_AddNode(t *testing.T) {
}

func TestServer_RemoveNode(t *testing.T) {
	s, m := newTestServerWithDefaults()
	ctx := context.Background()

	origVersion := m.ring.Version()

	nonExistent := &pb.Node{Id: 4242}

	r, err := s.RemoveNode(ctx, nonExistent)
	if err == nil {
		t.Errorf("RemoveNode(%#v) should have returned error (%s): %#v", nonExistent, err.Error(), r)
	}
	if err != nil {
		if r.Status {
			t.Errorf("RemoveNode(%#v) should have not returned true status: %#v", nonExistent, r)
		}
		if r.Version != origVersion {
			t.Errorf("RemoveNode(%#v) should have not returned new ring version: %#v", nonExistent, r)
		}
	}
	nodes, err := m.builder.Nodes().Filter([]string{"meta~=dummy1.*"})
	legitNode := &pb.Node{Id: nodes[0].ID()}

	r, err = s.RemoveNode(ctx, legitNode)
	if err != nil {
		t.Errorf("RemoveNode(%#v) should not have returned error (%s): %#v", legitNode, err.Error(), r)
	}
	if r.Version == origVersion || r.Status == false {
		t.Errorf("RemoveNode(%#v) should have returned new version and true status (%s): %#v", legitNode, err.Error(), r)
	}

	nodes, err = m.builder.Nodes().Filter([]string{"meta~=dummy1.*"})
	if len(nodes) != 0 {
		t.Errorf("RemoveNode(%#v) should have modified the builder but didn't", legitNode)
	}
}

func TestServer_ModNode(t *testing.T) {
}

func TestServer_SetConf(t *testing.T) {
}

func TestServer_SetActive(t *testing.T) {
}

func TestServer_GetVersion(t *testing.T) {
	s, m := newTestServerWithDefaults()
	ctx := context.Background()
	empty := &pb.EmptyMsg{}

	r, err := s.GetVersion(ctx, empty)
	if err != nil {
		t.Errorf("GetVersion() should not have rerturned an error: %s", err.Error())
	}

	if r.Version != m.ring.Version() {
		t.Errorf("GetVersion() returned ring version %d, expected %d", r.Version, m.ring.Version())
	}
}

func TestServer_GetGlobalConfig(t *testing.T) {
	s, m := newTestServerWithDefaults()
	ctx := context.Background()
	empty := &pb.EmptyMsg{}

	r, err := s.GetGlobalConfig(ctx, empty)
	if err != nil {
		t.Errorf("GetGlobalConfig() should not have rerturned an error: %s", err.Error())
	}
	if !bytes.Equal(r.Conf.Conf, m.builder.Conf()) {
		t.Errorf("GetGlobalConfig() returned config: %#v expected %#v", r.Conf.Conf, m.builder.Conf())
	}
}

func TestServer_SearchNodes(t *testing.T) {
}

func TestServer_GetNodeConfig(t *testing.T) {
}

func TestServer_GetRing(t *testing.T) {
	s, m := newTestServerWithDefaults()
	ctx := context.Background()
	empty := &pb.EmptyMsg{}
	rmsg, err := s.GetRing(ctx, empty)
	if rmsg.Version != m.ring.Version() {
		t.Errorf("GetRing() ring version was %d, expected %d", rmsg.Version, m.ring.Version())
	}
	if err != nil {
		t.Errorf("GetRing() returned unexpected error: %s", err.Error())
	}
	if !bytes.Equal(rmsg.Ring, *m.ringbytes) {
		log.Printf("%#v", rmsg)
		log.Printf("%d, %#v", m.ring.Version(), *m.ringbytes)
		t.Errorf("GetRing() returned ring bytes don't match expected")
	}
}

func TestServer_RegisterNode(t *testing.T) {
	s, _ := newTestServerWithDefaults()
	ctx := context.Background()
	badRequests := map[string]*pb.RegisterRequest{
		"Bad Network Interface": &pb.RegisterRequest{
			Hostname: "badnetiface.test.com",
			Addrs:    []string{"127.0.0.1/32", "192.168.2.2/32"},
			Tiers:    []string{"badnetiface.test.com", "zone1"},
		},
		"Duplicate server name and addr": &pb.RegisterRequest{
			Hostname: "server1",
			Addrs:    []string{"1.2.3.4/32", "127.0.0.1/32", "192.168.2.2/32"},
			Tiers:    []string{"server1", "zone1"},
		},
		"Bad tier": &pb.RegisterRequest{
			Hostname: "server42",
			Addrs:    []string{"10.0.0.42/32", "127.0.0.1/32", "192.168.2.2/32"},
			Tiers:    []string{""},
		},
	}

	for k, _ := range badRequests {
		r, err := s.RegisterNode(ctx, badRequests[k])
		if err == nil {
			t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have returned error because %s", badRequests[k], r, err.Error(), k)
		}
	}

	//now add a valid entry with the default strategy
	validRequest := &pb.RegisterRequest{
		Hostname: "server2",
		Addrs:    []string{"10.0.0.2/32", "127.0.0.1/32"},
		Tiers:    []string{"server2", "zone2"},
	}
	r, err := s.RegisterNode(ctx, validRequest)
	if err != nil {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have succeeded", validRequest, r, err.Error())
	}
	nodesbyaddr, _ := s.b.Nodes().Filter([]string{"address~=10.0.0.2"})
	nodesbymeta, _ := s.b.Nodes().Filter([]string{"meta~=server2.*"})
	if len(nodesbyaddr) != 1 || len(nodesbyaddr) != 1 {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have resulted with new ring entry. by addr (%v), by meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	if len(nodesbyaddr) == 1 && len(nodesbyaddr) == 1 {
		if r.Localid != nodesbyaddr[0].ID() || r.Localid != nodesbymeta[0].ID() {
			t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) localid should have matched id's in ring. by addr (%d), by meta (%d)", validRequest, r, err.Error(), nodesbyaddr[0].ID(), nodesbymeta[0].ID())
		}
	} else {
		t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) missing a entry by addr (%v) or meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	//verify default weight assignment
	if nodesbyaddr[0].Capacity() != 0 || nodesbyaddr[0].Active() == true {
		t.Errorf("RegisterNodes weight assignment strategy is default but found node with capacity (%d) and active (%v)", nodesbyaddr[0].Capacity(), nodesbyaddr[0].Active())
	}

	//test node already in ring
	validRequest.Reset()
	validRequest = &pb.RegisterRequest{
		Hostname: "server2",
		Addrs:    []string{"10.0.0.2/32", "127.0.0.1/32"},
		Tiers:    []string{"server2", "zone2"},
	}
	r, err = s.RegisterNode(ctx, validRequest)
	if err != nil {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have succeeded", validRequest, r, err.Error())
	}
	nodesbyaddr, _ = s.b.Nodes().Filter([]string{"address~=10.0.0.2"})
	nodesbymeta, _ = s.b.Nodes().Filter([]string{"meta~=server2.*"})
	if len(nodesbyaddr) != 1 || len(nodesbyaddr) != 1 {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should not have resulted with new ring entry. by addr (%v), by meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	if len(nodesbyaddr) == 1 && len(nodesbyaddr) == 1 {
		if r.Localid != nodesbyaddr[0].ID() || r.Localid != nodesbymeta[0].ID() {
			t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) localid should have matched id's in ring. by addr (%d), by meta (%d)", validRequest, r, err.Error(), nodesbyaddr[0].ID(), nodesbymeta[0].ID())
		} else {
			//verify default weight assignment
			if nodesbyaddr[0].Capacity() != 0 || nodesbyaddr[0].Active() != false {
				t.Errorf("RegisterNodes weight assignment strategy is default but found node with capacity (%d) and active (%v)", nodesbyaddr[0].Capacity(), nodesbyaddr[0].Active())
			}
		}
	} else {
		t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) missing a entry by addr (%v) or meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}

	//test fixed
	s.cfg.WeightAssignment = "fixed"
	validRequest.Reset()
	validRequest = &pb.RegisterRequest{
		Hostname: "server3",
		Addrs:    []string{"10.0.0.3/32", "127.0.0.1/32"},
		Tiers:    []string{"server3", "zone3"},
	}
	r, err = s.RegisterNode(ctx, validRequest)
	if err != nil {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have succeeded", validRequest, r, err.Error())
	}
	nodesbyaddr, _ = s.b.Nodes().Filter([]string{"address~=10.0.0.3"})
	nodesbymeta, _ = s.b.Nodes().Filter([]string{"meta~=server3.*"})
	if len(nodesbyaddr) != 1 || len(nodesbyaddr) != 1 {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should not have resulted with new ring entry. by addr (%v), by meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	if len(nodesbyaddr) == 1 && len(nodesbyaddr) == 1 {
		if r.Localid != nodesbyaddr[0].ID() || r.Localid != nodesbymeta[0].ID() {
			t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) localid should have matched id's in ring. by addr (%d), by meta (%d)", validRequest, r, err.Error(), nodesbyaddr[0].ID(), nodesbymeta[0].ID())
		} else {
			//verify default weight assignment
			if nodesbyaddr[0].Capacity() != 1000 || nodesbyaddr[0].Active() != true {
				t.Errorf("RegisterNodes weight assignment strategy is fixed but found node with capacity (%d) and active (%v)", nodesbyaddr[0].Capacity(), nodesbyaddr[0].Active())
			}
		}
	} else {
		t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) missing a entry by addr (%v) or meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}

	// test self
	s.cfg.WeightAssignment = "self"
	validRequest.Reset()
	validRequest = &pb.RegisterRequest{
		Hostname: "server4",
		Addrs:    []string{"10.0.0.4/32", "127.0.0.1/32"},
		Tiers:    []string{"server4", "zone4"},
	}
	r, err = s.RegisterNode(ctx, validRequest)
	if err != nil {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have succeeded", validRequest, r, err.Error())
	}
	nodesbyaddr, _ = s.b.Nodes().Filter([]string{"address~=10.0.0.4"})
	nodesbymeta, _ = s.b.Nodes().Filter([]string{"meta~=server4.*"})
	if len(nodesbyaddr) != 1 || len(nodesbyaddr) != 1 {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should not have resulted with new ring entry. by addr (%v), by meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	if len(nodesbyaddr) == 1 && len(nodesbyaddr) == 1 {
		if r.Localid != nodesbyaddr[0].ID() || r.Localid != nodesbymeta[0].ID() {
			t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) localid should have matched id's in ring. by addr (%d), by meta (%d)", validRequest, r, err.Error(), nodesbyaddr[0].ID(), nodesbymeta[0].ID())
		} else {
			//verify default weight assignment
			if nodesbyaddr[0].Capacity() != 1000 || nodesbyaddr[0].Active() != true {
				t.Errorf("RegisterNodes weight assignment strategy is self but found node with capacity (%d) and active (%v)", nodesbyaddr[0].Capacity(), nodesbyaddr[0].Active())
			}
		}
	} else {
		t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) missing a entry by addr (%v) or meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}

	//test manual
	s.cfg.WeightAssignment = "manual"
	validRequest.Reset()
	validRequest = &pb.RegisterRequest{
		Hostname: "server5",
		Addrs:    []string{"10.0.0.5/32", "127.0.0.1/32"},
		Tiers:    []string{"server5", "zone5"},
	}
	r, err = s.RegisterNode(ctx, validRequest)
	if err != nil {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should have succeeded", validRequest, r, err.Error())
	}
	nodesbyaddr, _ = s.b.Nodes().Filter([]string{"address~=10.0.0.5"})
	nodesbymeta, _ = s.b.Nodes().Filter([]string{"meta~=server5.*"})
	if len(nodesbyaddr) != 1 || len(nodesbyaddr) != 1 {
		t.Errorf("RegisterNode(ctx, %#v) (%#v, %s) should not have resulted with new ring entry. by addr (%v), by meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}
	if len(nodesbyaddr) == 1 && len(nodesbyaddr) == 1 {
		if r.Localid != nodesbyaddr[0].ID() || r.Localid != nodesbymeta[0].ID() {
			t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) localid should have matched id's in ring. by addr (%d), by meta (%d)", validRequest, r, err.Error(), nodesbyaddr[0].ID(), nodesbymeta[0].ID())
		} else {
			//verify default weight assignment
			if nodesbyaddr[0].Capacity() != 0 || nodesbyaddr[0].Active() != false {
				t.Errorf("RegisterNodes weight assignment strategy is manual but found node with capacity (%d) and active (%v)", nodesbyaddr[0].Capacity(), nodesbyaddr[0].Active())
			}
		}
	} else {
		t.Errorf("RegisterNode(ctx, %#v), (%#v, %s) missing a entry by addr (%v) or meta (%v)", validRequest, r, err.Error(), nodesbyaddr, nodesbymeta)
	}

}

func TestParseSlaveAddrs(t *testing.T) {
	slaves := []string{"1.1.1.1:8000", "2.2.2.2:8000"}
	rslaves := parseSlaveAddrs(slaves)
	if len(rslaves) != 2 {
		t.Errorf("parseSlaveAddrs(%#v), should have returned 2 RingSlaves but got: %#v", slaves, rslaves)
	}
}

func TestServer_ParseConfig(t *testing.T) {
}

func TestServer_LoadRingBuilderBytes(t *testing.T) {
}

func TestServer_RingBuilderPersisterFn(t *testing.T) {
	s, m := newTestServerWithDefaults()
	m.builder.SetConf([]byte("persisttest"))

	change := &RingChange{
		r: m.builder.Ring(),
		b: m.builder,
	}
	change.v = change.r.Version()
	tmpdir, err := ioutil.TempDir("", "rbpfntest")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	s.cfg.RingDir = tmpdir
	berr, rerr := s.ringBuilderPersisterFn(change, false)
	if berr != nil || rerr != nil {
		t.Errorf("ringBuilderPersisterFn(%v, false), should not have errored, got: %v, %v", change, berr, rerr)
	}
	berr, rerr = s.ringBuilderPersisterFn(change, true)
	if berr != nil || rerr != nil {
		t.Errorf("ringBuilderPersisterFn(%v, true), should not have errored, got: %v, %v", change, berr, rerr)
	}
	entries, err := ioutil.ReadDir(tmpdir)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	bnameExpected := fmt.Sprintf("%d-%s.builder", change.v, s.servicename)
	rnameExpected := fmt.Sprintf("%d-%s.ring", change.v, s.servicename)

	var bfound bool
	var rfound bool

	for _, entry := range entries {
		if entry.Name() == bnameExpected {
			bfound = true
		}
		if entry.Name() == rnameExpected {
			rfound = true
		}
	}

	if !bfound || !rfound {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Errorf("Couldn't find %s or %s file in tmpdir (%s): %#v", bnameExpected, rnameExpected, tmpdir, names)
	}

	bnameExpected = fmt.Sprintf("%s.builder", s.servicename)
	rnameExpected = fmt.Sprintf("%s.ring", s.servicename)

	bfound = false
	rfound = false

	for _, entry := range entries {
		if entry.Name() == bnameExpected {
			bfound = true
		}
		if entry.Name() == rnameExpected {
			rfound = true
		}
	}

	if !bfound || !rfound {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Errorf("Couldn't find %s or %s file in tmpdir (%s): %#v", bnameExpected, rnameExpected, tmpdir, names)
	}

	err = os.RemoveAll(tmpdir)
	if err != nil {
		t.Errorf(err.Error())
	}

}

func TestServer_ApplyRingChange(t *testing.T) {

	s, m := newTestServerWithDefaults()

	m.builder.SetConf([]byte("persisttest"))

	change := &RingChange{
		r: m.builder.Ring(),
		b: m.builder,
	}
	change.v = change.r.Version()

	err := s.applyRingChange(change)
	if err != nil {
		t.Errorf("applyRingChange(%v), should not have errored, got: %v", change, err)
	}

	m.persistBuilderErr = fmt.Errorf("persist builder oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have errored because of failed builder persist", change)
	}
	m.persistBuilderErr = nil

	m.persistRingErr = fmt.Errorf("persist ring oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have errored because of failed ring persist", change)
	}
	m.persistRingErr = nil

	m.bytesLoaderErr = fmt.Errorf("loader oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have failed because of loader err.", change)
	}

}

func TestServer_ValidNodeIP(t *testing.T) {
	s, _ := newTestServerWithDefaults()

	_, netblock, _ := net.ParseCIDR("10.0.0.0/24")
	s.netlimits = append(s.netlimits, netblock)

	loopback := "127.0.0.1/32"
	multicast := "224.0.0.1/32"
	inlimit := "10.0.0.1/32"
	badips := []string{
		loopback,
		multicast,
		"2.2.2.2/32",
	}
	for _, v := range badips {
		i, _, _ := net.ParseCIDR(v)
		if s.validNodeIP(i) {
			t.Errorf("validNodeIP(%s) should have been false but was true", v)
		}
	}
	log.Println(inlimit)
	okip, _, err := net.ParseCIDR(inlimit)
	if err != nil {
		panic(err)
	}
	if !s.validNodeIP(okip) {
		t.Errorf("validNodeIP(%s) should have been true but was false", okip)
	}
}

func TestServer_ValidTiers(t *testing.T) {
	s, _ := newTestServerWithDefaults()
	s.tierlimits = append(s.tierlimits, "*.zone")

	oktiers := []string{"localhost", "zone1"}
	exists := []string{"server1"}
	notiers := []string{}

	if !s.validTiers(oktiers) {
		t.Errorf("validTiers(%#v), should have been true", oktiers)
	}

	if s.validTiers(exists) {
		t.Errorf("validTiers(%#v), should have been false because host already exists", exists)
	}

	if s.validTiers(notiers) {
		t.Errorf("validTiers(%#v), should have been false because not enough tiers", notiers)
	}

}

func TestServer_NodeInRing(t *testing.T) {
	s, _ := newTestServerWithDefaults()

	if !s.nodeInRing("server1", []string{"1.2.3.4:56789"}) {
		t.Errorf("nodeInRing(server1, 1.2.3.4:56789), should have been true because node is in ring")
	}

	if !s.nodeInRing("server2", []string{"1.2.3.4:56789"}) {
		a := strings.Join([]string{s.r.Nodes()[0].Address(0)}, "|")
		log.Printf(".%s.", a)
		r, err := s.r.Nodes().Filter([]string{fmt.Sprintf("address~=%s", a)})
		log.Println(r)
		log.Println(err)
		t.Errorf("nodeInRing(server2, 1.2.3.4:56789), should have been true because ip is already in ring")
	}

	if s.nodeInRing("server2", []string{"1.2.3.5:56789"}) {
		t.Errorf("nodeInRing(server2, 1.2.3.5:56789), should have been false because node is not in ring")
	}
}
