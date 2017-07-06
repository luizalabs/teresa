package deploy

import (
	"bufio"
	"bytes"
	"io"

	"google.golang.org/grpc"

	"github.com/luizalabs/teresa-api/models/storage"
	dpb "github.com/luizalabs/teresa-api/pkg/protobuf/deploy"
)

type Service struct {
	ops Operations
}

func (s *Service) Make(stream dpb.Deploy_MakeServer) error {
	var appName, description string
	content := new(bytes.Buffer)

	ctx := stream.Context()
	u := ctx.Value("user").(*storage.User)

	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if info := in.GetInfo(); info != nil {
			appName = info.App
			description = info.Description
		}
		if data := in.GetFile(); data != nil {
			content.Write(data.Chunk)
		}
	}

	rs := bytes.NewReader(content.Bytes())
	rc, err := s.ops.Deploy(u, appName, rs, description)
	if err != nil {
		return err
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		if err := stream.Send(&dpb.DeployResponse{Text: scanner.Text()}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	dpb.RegisterDeployServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
