package object

type Environment struct {
	names map[string]Object
}

func NewEnvironment() *Environment {
	return &Environment{
		make(map[string]Object),
	}
}

func (env *Environment) Get(name string) (Object, bool) {
	obj, ok := env.names[name]
	return obj, ok
}

func (env *Environment) Set(name string, obj Object) Object {
	env.names[name] = obj
	return obj
}
