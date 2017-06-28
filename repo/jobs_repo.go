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

	// Initialize the deluge (potentially long, so we do that outside of mutex-lock)
	start := time.Now()
	dlg, err := core.NewDeluge(id, program)
	if err != nil {
		return nil, err
	}
	log.Infof("Deluge initialized in %s", time.Now().Sub(start).String())

	jr.jobsMutex.Lock()
	defer jr.jobsMutex.Unlock()

	// Checks that ID is not already taken before inserting it in the repository
	if _, ok := jr.jobs[id]; ok {
		return nil, errors.New(fmt.Sprintf("Cannot create job with id '%s'. A job with this id already exists.", id))
	}
	jr.jobs[id] = dlg
	return dlg, nil
}

func (jr *JobsRepository) Get(id string) (*core.Deluge, bool) {
	jr.jobsMutex.Lock()
	defer jr.jobsMutex.Unlock()
	dlg, ok := jr.jobs[id]
	return dlg, ok
}

func (jr *JobsRepository) GetAll() []*core.Deluge {
	jr.jobsMutex.Lock()
	defer jr.jobsMutex.Unlock()
	all := make([]*core.Deluge, 0, len(jr.jobs))
	for _, v := range jr.jobs {
		all = append(all, v)
	}
	return all
}

func (jr *JobsRepository) Delete(id string) bool {
	jr.jobsMutex.Lock()
	defer jr.jobsMutex.Unlock()
	if _, ok := jr.jobs[id]; ok {
		delete(jr.jobs, id)
		return true
	}
	return false
}
