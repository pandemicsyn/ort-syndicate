package syndicate

import (
	"fmt"

	pb "github.com/pandemicsyn/syndicate/api/proto"
	"golang.org/x/net/context"
)

//GetNodeSoftwareVersion asks a managed node for its running software version
func (s *Server) GetNodeSoftwareVersion(c context.Context, n *pb.Node) (*pb.NodeSoftwareVersion, error) {
	s.RLock()
	defer s.RUnlock()
	node := s.r.Node(n.Id)
	if node == nil {
		return &pb.NodeSoftwareVersion{}, fmt.Errorf("Node %d not found", n.Id)
	}
	version, err := s.managedNodes[n.Id].GetSoftwareVersion()
	return &pb.NodeSoftwareVersion{Version: version}, err
}

//GetNodeSoftwareVersion asks a managed node for its running software version
func (s *Server) NodeUpgradeSoftwareVersion(c context.Context, n *pb.NodeUpgrade) (*pb.NodeUpgradeStatus, error) {
	s.RLock()
	defer s.RUnlock()
	node := s.r.Node(n.Id)
	if node == nil {
		return &pb.NodeUpgradeStatus{Status: false, Msg: ""}, fmt.Errorf("Node %d not found", n.Id)
	}
	status, err := s.managedNodes[n.Id].UpgradeSoftwareVersion(n.Version)
	return &pb.NodeUpgradeStatus{Status: status, Msg: ""}, err
}
