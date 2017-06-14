package repo

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge/deluge"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type JobsRepository struct {
	jobs      map[string]*deluge.Deluge
	jobsMutex *sync.Mutex
}

var Jobs = NewJobsRepository()

func NewJobsRepository() *JobsRepository {
	return &JobsRepository{
		jobs:      make(map[string]*deluge.Deluge),
		jobsMutex: &sync.Mutex{},
	}
}

func (jr *JobsRepository) Create(program *ast.Program) *deluge.Deluge {
	start := time.Now()
	id := uuid.NewV4().String()
	dlg := deluge.NewDeluge(id, program)
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())

	jr.jobsMutex.Lock()
	jr.jobs[id] = dlg
	jr.jobsMutex.Unlock()

	return dlg
}

func (jr *JobsRepository) Get(id string) (*deluge.Deluge, bool) {
	jr.jobsMutex.Lock()
	dlg, ok := jr.jobs[id]
	jr.jobsMutex.Unlock()
	return dlg, ok
}

func (jr *JobsRepository) GetAll() []*deluge.Deluge {
	jr.jobsMutex.Lock()
	all := make([]*deluge.Deluge, 0, len(jr.jobs))
	for _, v := range jr.jobs {
		all = append(all, v)
	}
	jr.jobsMutex.Unlock()
	return all
}

func (jr *JobsRepository) Delete(id string) bool {
	jr.jobsMutex.Lock()
	if _, ok := jr.jobs[id]; ok {
		delete(jr.jobs, id)
		return true
	}
	jr.jobsMutex.Unlock()
	return false
}
