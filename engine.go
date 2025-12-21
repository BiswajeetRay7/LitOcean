package main

import "sync"

type Engine struct {
	Paused bool
	Mu     sync.Mutex
}

func (e *Engine) Pause() {
	e.Mu.Lock()
	e.Paused = true
	e.Mu.Unlock()
}

func (e *Engine) Resume() {
	e.Mu.Lock()
	e.Paused = false
	e.Mu.Unlock()
}

func (e *Engine) Wait() {
	for {
		e.Mu.Lock()
		p := e.Paused
		e.Mu.Unlock()
		if !p {
			return
		}
	}
}
