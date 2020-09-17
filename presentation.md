% The Go Programming Language
% Davi Diorio Mendes (d.mendes@sidi.org.br)
% 12/10/2020

# The Go Programming Language

## Agenda

- Creators
- What is Go?
- Why?
- Target
- How?
- OO Features
- Cases
- Live Coding

---

# Creators

*Ken Thompson*

- Unix, Plan9 (Bell Labs)
- UTF-8
- Linguagem B

*Rob Pike*

- Unix, Plan9 (Bell Labs)
- UTF-8
- Newsqueak, Limbo

*Robert Griesemmer*

- HotSpot (JVM)
- Code generation for Google V8
- Sawzall (linguagem)

---

# What is Go?

General purpose but created to address
the pain of a server application team.

Designed to address Google's problems
with compilation time, documentation
and maintainability.

Structured paradigm.

Statically typed.

Focused on concurrent programming.

---

# Why?

The programming language was no more a
tool and become another challenge for
the project.

Languages for efficient execution have
high build times and high complexity.

Complex languages leads to each developer
gets used to a subset of the language.

Poor documentation quality.

Languages were not designed for modern
computers since the beginning.

---

# Target

Efficient Compilation

Efficient Execution

Easy of Programming

---

# How?

Very neat grammar with orthogonal elements.

Derived from C family.

Transfer most of the complexity from
the developer to the language runtime.

Channels and goroutines for efficient
and easy concurrent programming.

The import statement is part of the language.

Unused imports are a compilation error.

The dependencies are compiled once. Further
imports just reads the package contract
at the beggining of the object file.

---

# OO Features

Bind methods to types.

Implicit interfaces.

Composition instead of type hierarchy.

---

# Cases

## Fox Sports

People are using more and more streaming
services over television.

Users were experiencing instability during
super bowl.

Refactoring the streaming service from Node
to Go reduced cloud infra costs and increased
traffic capacity.

A load test with 30k requests were crashing
and rebooting the former service due to high
CPU usage. Go services kept it constant at
12% and 20MB of memory usage.

https://x-team.com/blog/fox-sports/

---

# Cases

## Uber Frontless

Sharding layer for a high availability postgres
architecture. Formerly named Schemaless was written
in Python but as Uber adoption grew the intense
usage of Schemaless lead to hagh latency.

Frontless is the same Schemaless API but written
in Go. The migration reduced latency in 85%.

During a load test with constant rate of requests
the Frontless CPU usage was always below 15% while
the Schemaless CPU usage was 100%.

The high performance allowed Uber to reduce its
instances of the sharding layer and still deliver
a high quality service.

https://eng.uber.com/schemaless-rewrite/

---

# Cases

https://go.dev/solutions

---

# Questions?

---

# Live Coding

---

# References

https://go.dev/

https://talks.golang.org/2012/splash.article

https://golang.org/doc/faq
