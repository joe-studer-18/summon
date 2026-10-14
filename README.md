# Summon

A magical Dependency Injection framework for Go.

## Why

There's little question that Dependency Injection is the correct way to architect your object deps. Its
composable, testable, and best of all, obvious. And you can always just stop there - no need 
for a framework. But, ya know, frameworks are nice - so what if we built one?

There are some great options out there for DI frameworks in Go - but they all felt like they were just
missing a little bit of that magic that I'd come to expect from tools like Guice, back when I was writing 
Java. 

I make no illusions that `summon` would make the authors of Golang happy - I haven't even looked to see what
their official stance is. So, if by some ill-fated fortune you're one of those maintainers, and reading this, 
hey I love your work, but imma do this regardless.

## Usage

### Registering your type with Summon

```golang
type Foo struct{}

func (f Foo) Summon() Foo {
    return Foo{}
}

type Bar struct{
    foo Foo
}

// Circle will find Foo and provide it _like magic_
func (b Bar) Summon(f Foo) Bar {
    return Bar{f}
}
```

### Summoning your type

```golang
// 1. Create a circle
circle := NewCircle()

// 2. Summon what you need!
s := circle.Summon[Foo]()
```

### Binding your type

At times, you may need a more advanced incantation - perhaps to 
summon an object you don't own, or to provide an implementation of
an interface. In these cases, you'll want to `Bind` the
implementation to the interface.

For example:

```golang
type AnInterface interface {
	Do() string
}

type AnImplementation struct{}

var _ = Binds[AnInterface, AnImplementation]()

func (a AnImplementation) Summon() AnImplementation {
	return AnImplementation{}
}

func (a AnImplementation) Do() string {
	return "AnImplementation"
}
```

From there, summoning your implementation is no different:

```golang
circle := NewCircle()

a := circle.Summon[AnInterface]()
```

### Singletons

Singletons get a special case - instead of using the normal `Summon`, you now
need `SummonSingleton`. Take the snippet below:

```golang
type Singleton struct {}

func (*Singleton) SummonSingleton() *Singleton {
    return new(Singleton{})
}
```

Note that we're using a pointer here - that's required! A `SummonSingleton`
that provides a direct object will throw an error when you attempt
to `circle.Summon` it.

## Contributors Guide

### AI Stance

I'm not accepting AI code, and I won't be using it in this project. I _will_ use AI to learn - it's a solid 
search tool. But every commit, ever line, is _artisinal_. I'm doing this for me.

### Raising a PR

Before raising a PR, please do the following:

* Open and self-assign an issue describing the change you'd like to make
* Add appropriate test coverage to your change. If it's testable, it's best to have coverage on it