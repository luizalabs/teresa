package app

type LogOptions struct {
	Lines   int64
	Follow  bool
	PodName string
}

type PodListOptions struct {
	PodName string
}
