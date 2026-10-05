# digo

A Dependency Injection framework for Golang - thus, digo.

## Why

There's little question that Dependency Injection is the correct way to architect your object deps. Its
composable, testable, and best of all, obvious. And you can always just stop there - no need 
for a framework. But, ya know, frameworks are nice - so what if we built one?

There are some great options out there for DI frameworks in Go - but they all felt like they were just
missing a little bit of that magic that I'd come to expect from tools like Guice, back when I was writing 
Java. 

I make no illusions that `digo` would make the authors of Golang happy - I haven't even looked to see what
their official stance is. So, if by some ill-fated fortune you're one of those maintainers, and reading this, 
hey I love your work, but imma do this regardless.

## AI Stance

I'm not accepting AI code, and I won't be using it in this project. I _will_ use AI to learn - it's a solid semantic 
search tool. But every commit, ever line, is _artisinal_. I'm doing this for me.

## Open Features

These are features that are needed to get parity with other DI frameworks, while keeping the magic.

* Support for interfaces
* Tests!
* Singletons
* Caching/Performance optimizations
* Name based bindings, so you can provide multiple versions of a thing
* Sensible path for third-party types (probably factories)