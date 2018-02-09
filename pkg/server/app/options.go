package app

type LogOptions struct {
	Lines    int64
	Follow   bool
	PodName  string
	Previous bool
}

type PodListOptions struct {
	PodName string
}
