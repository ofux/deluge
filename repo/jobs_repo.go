package repo

import (
	"github.com/ofux/deluge/core"
	"sync"
)

type RunningJob struct {
	ID             string
	RunnableDeluge *core.RunnableDeluge
}

type RunningJobs struct {
	jobs  map[string]*RunningJob
	mutex *sync.Mutex
}

var Jobs = NewJobsRepository()

func NewJobsRepository() *RunningJobs {
	return &RunningJobs{
		jobs:  make(map[string]*RunningJob),
		mutex: &sync.Mutex{},
	}
}

func (jr *RunningJobs) Store(job *RunningJob) {
	jr.mutex.Lock()
	defer jr.mutex.Unlock()
	jr.jobs[job.ID] = job
}

func (jr *RunningJobs) Get(id string) (*RunningJob, bool) {
	jr.mutex.Lock()
	defer jr.mutex.Unlock()
	job, ok := jr.jobs[id]
	return job, ok
}

func (jr *RunningJobs) GetAll() []*RunningJob {
	jr.mutex.Lock()
	defer jr.mutex.Unlock()
	all := make([]*RunningJob, 0, len(jr.jobs))
	for _, v := range jr.jobs {
		all = append(all, v)
	}
	return all
}
