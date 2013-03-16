// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lab provides a global module register.
package lab

// Module is a part of your program.
type Module interface{}

// Initer modules are initialized in registration order on lab start.
type Initer interface {
	Init()
}

// Runner modules are run in a new goroutine on lab start.
type Runner interface {
	Run()
}

var (
	lock bool
	list []Module
	mods = map[string]Module{}
)

// Register registers a module by name. It must be called before the lab is started.
func Register(name string, mod Module) {
	if lock {
		panic("lab was started")
	}
	list = append(list, mod)
	mods[name] = mod
}

// Mod returns a registered module with name or nil.
func Mod(name string) Module {
	return mods[name]
}

// All returns all registered modules.
func All() []Module {
	m := make([]Module, len(list))
	copy(m, list)
	return m
}

// Start locks the lab and initializes and runs all modules.
func Start() {
	if lock {
		panic("lab was started")
	}
	lock = true
	for _, m := range list {
		if i, ok := m.(Initer); ok {
			i.Init()
		}
	}
	for _, m := range list {
		if r, ok := m.(Runner); ok {
			go r.Run()
		}
	}
}
