package object

type Environment struct {
	names map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		names: make(map[string]Object),
	}
}

func WrappedEnvironment(outerEnv *Environment) *Environment {
	inner := NewEnvironment()
	inner.outer = outerEnv
	return inner
}

func (env *Environment) Get(name string) (Object, bool) {
	obj, ok := env.names[name]
	if !ok && env.outer != nil {
		obj, ok = env.outer.Get(name)
	}
	return obj, ok
}

func (env *Environment) Set(name string, obj Object) Object {
	env.names[name] = obj
	return obj
}
