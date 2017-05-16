package team

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrTeamAlreadyExists = status.Errorf(codes.AlreadyExists, "Team already exists")
