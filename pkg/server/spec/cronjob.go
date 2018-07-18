package spec

const defaultJobHistoryLimit = 3

type CronJob struct {
	Deploy
	Schedule                   string
	SuccessfulJobsHistoryLimit int32
	FailedJobsHistoryLimit     int32
}

type CronJobBuilder struct {
	d        Deploy
	schedule string
}

func (b *CronJobBuilder) WithDescription(description string) *CronJobBuilder {
	b.d.Description = description
	return b
}

func (b *CronJobBuilder) WithPod(p *Pod) *CronJobBuilder {
	b.d.Pod = *p
	return b
}

func (b *CronJobBuilder) WithSchedule(s string) *CronJobBuilder {
	b.schedule = s
	return b
}

func (b *CronJobBuilder) Build() *CronJob {
	return &CronJob{
		Deploy:                     b.d,
		Schedule:                   b.schedule,
		SuccessfulJobsHistoryLimit: defaultJobHistoryLimit,
		FailedJobsHistoryLimit:     defaultJobHistoryLimit,
	}
}

func NewCronJobBuilder(slugURL string) *CronJobBuilder {
	d := Deploy{
		SlugURL:     slugURL,
		MatchLabels: make(Labels),
	}
	return &CronJobBuilder{d: d}
}
