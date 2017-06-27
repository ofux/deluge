package repo

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type JobsRepository struct {
	jobs      map[string]*core.Deluge
	jobsMutex *sync.Mutex
}

var Jobs = NewJobsRepository()

func NewJobsRepository() *JobsRepository {
	return &JobsRepository{
		jobs:      make(map[string]*core.Deluge),
		jobsMutex: &sync.Mutex{},
	}
}

func (jr *JobsRepository) Create(program *ast.Program) (*core.Deluge, error) {
	return jr.CreateWithID(program, uuid.NewV4().String())
}

func (jr *JobsRepository) CreateWithID(program *ast.Program, id string) (*core.Deluge, error) {
	if _, ok := jr.jobs[id]; ok {
		return nil, errors.New(fmt.Sprintf("Cannot create job with id '%s'. A job with this id already exists.", id))
	}

	start := time.Now()
	dlg := core.NewDeluge(id, program)
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())

	jr.jobsMutex.Lock()
	jr.jobs[id] = dlg
	jr.jobsMutex.Unlock()

	return dlg, nil
}

func (jr *JobsRepository) Get(id string) (*core.Deluge, bool) {
	jr.jobsMutex.Lock()
	dlg, ok := jr.jobs[id]
	jr.jobsMutex.Unlock()
	return dlg, ok
}

func (jr *JobsRepository) GetAll() []*core.Deluge {
	jr.jobsMutex.Lock()
	all := make([]*core.Deluge, 0, len(jr.jobs))
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
