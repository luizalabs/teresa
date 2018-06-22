package build

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/luizalabs/teresa/pkg/goutil"
	bpb "github.com/luizalabs/teresa/pkg/protobuf/build"
	"github.com/luizalabs/teresa/pkg/server/database"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

const KeepAliveMessage = "\u200B" // Zero width space

type Service struct {
	ops              Operations
	keepAliveTimeout time.Duration
}

func (s *Service) Make(stream bpb.Build_MakeServer) error {
	var appName, name string
	var runApp bool

	content := new(bytes.Buffer)

	ctx := stream.Context()
	u := ctx.Value("user").(*database.User)

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
			name = info.Name
			runApp = info.Run
		}
		if data := in.GetFile(); data != nil {
			content.Write(data.Chunk)
		}
	}

	rs := bytes.NewReader(content.Bytes())
	rc, errChan := s.ops.Create(ctx, appName, name, u, rs, runApp)
	if rc == nil {
		return <-errChan
	}
	defer rc.Close()

	buildMsgs, buildErrCh := goutil.LineGenerator(rc)
	return s.streamMsg(ctx, buildMsgs, errChan, buildErrCh, func(msg string) error {
		return stream.Send(&bpb.BuildResponse{Text: msg})
	})
}

func (s *Service) List(ctx context.Context, req *bpb.ListRequest) (*bpb.ListResponse, error) {
	u := ctx.Value("user").(*database.User)

	items, err := s.ops.List(req.AppName, u)
	if err != nil {
		return nil, err
	}

	res := &bpb.ListResponse{Builds: make([]*bpb.ListResponse_Build, len(items))}
	for i := range items {
		res.Builds[i] = &bpb.ListResponse_Build{
			Name:         items[i].Name,
			LastModified: items[i].LastModified.String(),
		}
	}

	return res, nil
}

func (s *Service) Run(req *bpb.RunRequest, stream bpb.Build_RunServer) error {
	ctx := stream.Context()
	u := ctx.Value("user").(*database.User)

	rc, errChan := s.ops.Run(ctx, req.AppName, req.Name, u)
	if rc == nil {
		return <-errChan
	}
	defer rc.Close()

	runMsgs, runErrChan := goutil.LineGenerator(rc)
	return s.streamMsg(ctx, runMsgs, errChan, runErrChan, func(msg string) error {
		return stream.Send(&bpb.RunResponse{Text: msg})
	})
}

func (s *Service) streamMsg(ctx context.Context, msgsChan <-chan string, errChan, msgsErrChan <-chan error, sendFn func(string) error) error {
	var msg string
	for {
		select {
		case <-time.After(s.keepAliveTimeout):
			msg = KeepAliveMessage
		case err := <-errChan:
			return err
		case err := <-msgsErrChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-msgsChan:
			if !ok {
				return nil
			}
			msg = fmt.Sprintf("%s\n", m)
		}
		if err := sendFn(msg); err != nil {
			return err
		}
	}
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	bpb.RegisterBuildServer(grpcServer, s)
}

func NewService(ops Operations, keepAliveTimeout time.Duration) *Service {
	return &Service{ops: ops, keepAliveTimeout: keepAliveTimeout}
}
