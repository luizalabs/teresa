package client

import (
	"sort"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
)

type ByKey []*appb.InfoResponse_EnvVar

func (s ByKey) Len() int {
	return len(s)
}

func (s ByKey) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByKey) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func SortEnvsByKey(envs []*appb.InfoResponse_EnvVar) {
	sort.Sort(ByKey(envs))
}
