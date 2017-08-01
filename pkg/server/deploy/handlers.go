package deploy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"

	"github.com/luizalabs/teresa-api/models/storage"
	dpb "github.com/luizalabs/teresa-api/pkg/protobuf/deploy"
)

const (
	keepAliveMessage = "\u200B" // Zero width space
)

type Options struct {
	KeepAliveTimeout     time.Duration `split_words:"true" default:"30s"`
	RevisionHistoryLimit int           `split_words:"true" default:"5"`
	SlugBuilderImage     string        `split_words:"true" default:"luizalabs/slugbuilder:v2.4.9"`
	SlugRunnerImage      string        `split_words:"true" default:"luizalabs/slugrunner:v2.2.4"`
}

type Service struct {
	ops     Operations
	options *Options
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
	rc, err := s.ops.Deploy(u, appName, rs, description, s.options)
	if err != nil {
		return err
	}
	defer rc.Close()

	deployMsgs := channelFromReader(rc)
	var msg string

	for {
		select {
		case <-time.After(s.options.KeepAliveTimeout):
			msg = keepAliveMessage
		case m, ok := <-deployMsgs:
			if !ok {
				return nil
			}
			msg = m
		}

		if err := stream.Send(&dpb.DeployResponse{Text: msg}); err != nil {
			return err
		}
	}
}

func channelFromReader(r io.Reader) <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			c <- fmt.Sprintln(scanner.Text())
		}
	}()

	return c
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	dpb.RegisterDeployServer(grpcServer, s)
}

func NewService(ops Operations, options *Options) *Service {
	return &Service{ops: ops, options: options}
}
