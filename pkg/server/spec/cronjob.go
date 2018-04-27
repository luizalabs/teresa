package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	defaultJobHistoryLimit = 3
)

type CronJob struct {
	Deploy
	Schedule                   string
	SuccessfulJobsHistoryLimit int32
	FailedJobsHistoryLimit     int32
}

func NewCronJob(description, slugURL, schedule string, imgs *Images, a *app.App, fs storage.Storage, args ...string) *CronJob {
	ps := NewPod(
		a.Name,
		"",
		imgs.SlugRunner,
		a,
		map[string]string{
			"APP":      a.Name,
			"SLUG_URL": slugURL,
			"SLUG_DIR": slugVolumeMountPath,
		},
		fs,
	)
	ps.Containers[0].Args = args
	ps.Containers[0].VolumeMounts = []*VolumeMounts{newSlugVolumeMount()}
	ps.InitContainers = newInitContainers(slugURL, imgs.SlugStore, a, fs)

	ds := Deploy{
		Description: description,
		SlugURL:     slugURL,
		Pod:         *ps,
	}

	cs := &CronJob{
		Deploy:                     ds,
		Schedule:                   schedule,
		SuccessfulJobsHistoryLimit: defaultJobHistoryLimit,
		FailedJobsHistoryLimit:     defaultJobHistoryLimit,
	}

	return cs
}
