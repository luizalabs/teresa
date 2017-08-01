package deploy

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrPodRunFail            = status.Errorf(codes.Unknown, "Run command returned a non zero value")
	ErrBuildFail             = status.Errorf(codes.Unknown, "Build returned a non zero value")
	ErrReleaseFail           = status.Errorf(codes.Unknown, "Release command returned a non zero value")
	ErrInvalidTeresaYamlFile = status.Errorf(codes.InvalidArgument, "Invalid Teresa Yaml file")
)
