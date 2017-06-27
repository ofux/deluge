package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) getWithEnv(name string) (*Environment, Object, bool) {
	env := e
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		env, obj, ok = e.outer.getWithEnv(name)
	}
	return env, obj, ok
}

func (e *Environment) Get(name string) (Object, bool) {
	_, obj, ok := e.getWithEnv(name)
	return obj, ok
}

func (e *Environment) Add(name string, val Object) bool {
	_, ok := e.store[name]
	if ok {
		return false
	}
	e.store[name] = val
	return true
}

func (e *Environment) Set(name string, val Object) bool {
	env, _, ok := e.getWithEnv(name)
	if !ok {
		return false
	}
	env.store[name] = val
	return true
}
