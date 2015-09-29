package cmdctrl

import (
	"net"

	pb "github.com/pandemicsyn/ort-syndicate/api/cmdctrl"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CmdCtrl interface {
	Start() (err error)
	Stop() (err error)
	Reload() (err error)
	Restart() (err error)
	RingUpdate(version uint64, ringBytes []byte) (oldversion, newversion uint64)
	Stats() (encoded []byte)
}

type CCServer struct {
	cmdctrl CmdCtrl
	cfg     *ConfigOpts
}

type ConfigOpts struct {
	ListenAddress string
	CertFile      string
	KeyFile       string
	UseTLS        bool
}

func NewCCServer(c CmdCtrl, cfg *ConfigOpts) *CCServer {
	return &CCServer{
		cmdctrl: c,
		cfg:     cfg,
	}
}

func (c *CCServer) Serve() error {
	l, err := net.Listen("tcp", c.cfg.ListenAddress)
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	if c.cfg.UseTLS {
		creds, err := credentials.NewServerTLSFromFile(c.cfg.CertFile, c.cfg.KeyFile)
		if err != nil {
			return err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	s := grpc.NewServer(opts...)
	pb.RegisterCmdCtrlServer(s, c)
	return s.Serve(l)
}

func (cc *CCServer) RingUpdate(c context.Context, r *pb.Ring) (*pb.RingUpdateResult, error) {
	res := pb.RingUpdateResult{}
	res.Oldversion, res.Newversion = cc.cmdctrl.RingUpdate(r.Version, r.Ring)
	return &res, nil

}

func (cc *CCServer) Restart(c context.Context, r *pb.EmptyMsg) (*pb.StatusMsg, error) {
	err := cc.cmdctrl.Restart()
	if err != nil {
		return &pb.StatusMsg{Status: false, Msg: err.Error()}, nil
	}
	return &pb.StatusMsg{Status: true, Msg: ""}, nil
}

func (cc *CCServer) Start(c context.Context, r *pb.EmptyMsg) (*pb.StatusMsg, error) {
	err := cc.cmdctrl.Start()
	if err != nil {
		return &pb.StatusMsg{Status: false, Msg: err.Error()}, nil
	}
	return &pb.StatusMsg{Status: true, Msg: ""}, nil
}

func (cc *CCServer) Stop(c context.Context, r *pb.EmptyMsg) (*pb.StatusMsg, error) {
	err := cc.cmdctrl.Stop()
	if err != nil {
		return &pb.StatusMsg{Status: false, Msg: err.Error()}, nil
	}
	return &pb.StatusMsg{Status: true, Msg: ""}, nil
}

func (cc *CCServer) Reload(c context.Context, r *pb.EmptyMsg) (*pb.StatusMsg, error) {
	err := cc.cmdctrl.Reload()
	if err != nil {
		return &pb.StatusMsg{Status: false, Msg: err.Error()}, nil
	}
	return &pb.StatusMsg{Status: true, Msg: ""}, nil
}

func (cc *CCServer) Stats(c context.Context, r *pb.EmptyMsg) (*pb.StatsMsg, error) {
	return &pb.StatsMsg{Statsjson: cc.cmdctrl.Stats()}, nil
}
