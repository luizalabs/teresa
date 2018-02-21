package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type CronJob struct {
	Deploy
	Schedule string
}

func NewCronJob(image, description, slugURL, schedule string, a *app.App, fs storage.Storage, args ...string) *CronJob {
	ps := NewPod(
		a.Name,
		image,
		a,
		map[string]string{
			"APP":             a.Name,
			"SLUG_URL":        slugURL,
			"BUILDER_STORAGE": fs.Type(),
		},
		fs,
	)
	ps.Args = args

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
