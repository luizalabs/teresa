package app

import "sort"

type ByKey []*EnvVar

func (s ByKey) Len() int {
	return len(s)
}

func (s ByKey) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByKey) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func sortEnvsByKey(envs []*EnvVar) {
	sort.Sort(ByKey(envs))
}
