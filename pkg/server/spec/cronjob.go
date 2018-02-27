package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type CronJob struct {
	Deploy
	Schedule string
}

func NewCronJob(description, slugURL, schedule string, imgs *SlugImages, a *app.App, fs storage.Storage, args ...string) *CronJob {
	ps := NewPod(
		a.Name,
		imgs.Runner,
		a,
		map[string]string{
			"APP":      a.Name,
			"SLUG_URL": slugURL,
			"SLUG_DIR": slugVolumeMountPath,
		},
		fs,
	)
	ps.Args = args
	ps.VolumeMounts = []*VolumeMounts{newSlugVolumeMount()}
	ps.InitContainers = newInitContainers(slugURL, imgs.Store, a, fs)

	ds := Deploy{
		Description: description,
		SlugURL:     slugURL,
		Pod:         *ps,
	}

	cs := &CronJob{
		Deploy:   ds,
		Schedule: schedule,
	}

	return cs
}
