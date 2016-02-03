package syndicate

import (
	"fmt"
	"log"
	"testing"

	"github.com/gholt/ring"
)

func TestWithRingBuilderPersister(t *testing.T) {
}

func TestWithRingBuilderBytesLoader(t *testing.T) {
}

func TestNewServer(t *testing.T) {
}

func TestServer_AddNode(t *testing.T) {
}

func TestServer_RemoveNode(t *testing.T) {
}

func TestServer_ModNode(t *testing.T) {
}

func TestServer_SetConf(t *testing.T) {
}

func TestServer_SetActive(t *testing.T) {
}

func TestServer_GetVersion(t *testing.T) {
}

func TestServer_GetGlobalConfig(t *testing.T) {
}

func TestServer_SearchNodes(t *testing.T) {
}

func TestServer_GetNodeConfig(t *testing.T) {
}

func TestServer_GetRing(t *testing.T) {
}

func TestServer_RegisterNode(t *testing.T) {
}

func TestParseSlaveAddrs(t *testing.T) {
}

func TestServer_ParseConfig(t *testing.T) {
}

func TestServer_LoadRingBuilderBytes(t *testing.T) {
}

func TestServer_RingBuilderPersisterFn(t *testing.T) {
}

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

func newTestServerWithDefaults() (*Server, *MockRingBuilderThings) {
	b := ring.NewBuilder(64)
	b.SetReplicaCount(3)
	b.AddNode(true, 1, []string{"server1", "zone1"}, []string{"1.2.3.4:56789"}, "Meta One", []byte("Conf"))
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
		managedNodes: make(map[uint64]ManagedNode),
		slaves:       make([]*RingSlave, 0),
		changeChan:   make(chan *changeMsg, 1),
	}
	s := newTestServer(&Config{}, "test", mock)
	return s, mock
}

func newTestServer(cfg *Config, servicename string, mockinfo *MockRingBuilderThings) *Server {

	s := &Server{}
	s.cfg = cfg
	s.servicename = servicename
	//s.parseConfig()

	s.rbPersistFn = mockinfo.Persist
	s.rbLoaderFn = mockinfo.BytesLoader
	s.b = mockinfo.builder
	s.r = mockinfo.ring
	s.rb = mockinfo.ringbytes
	s.bb = mockinfo.builderbytes
	s.managedNodes = make(map[uint64]ManagedNode)
	s.slaves = mockinfo.slaves
	s.changeChan = mockinfo.changeChan
	return s
}

func TestServer_ApplyRingChange(t *testing.T) {

	s, m := newTestServerWithDefaults()
	change := &RingChange{v: 42}
	err := s.applyRingChange(change)
	if err != nil {
		t.Errorf("applyRingChange(%v), should not have errored, got: %v", change, err)
	}

	m.persistBuilderErr = fmt.Errorf("oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have errored because of failed builder persist", change)
	}
	m.persistBuilderErr = nil

	m.persistRingErr = fmt.Errorf("oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have errored because of failed ring persist", change)
	}
	m.persistRingErr = nil

	m.bytesLoaderErr = fmt.Errorf("oops")
	err = s.applyRingChange(change)
	if err == nil {
		t.Errorf("applyRingChange(%v), should have failed because of loader err.", change)
	}

}

func TestServer_ValidNodeIP(t *testing.T) {
}

func TestServer_ValidTiers(t *testing.T) {
}

func TestServer_NodeInRing(t *testing.T) {
}
