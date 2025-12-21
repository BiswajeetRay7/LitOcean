package main

import "sync"

type Engine struct {
	Paused bool
	mu     sync.Mutex
}

func (e *Engine) Pause() {
	e.mu.Lock()
	e.Paused = true
	e.mu.Unlock()
}

func (e *Engine) Resume() {
	e.mu.Lock()
	e.Paused = false
	e.mu.Unlock()
}

func (e *Engine) Wait() {
	for {
		e.mu.Lock()
		p := e.Paused
		e.mu.Unlock()
		if !p {
			return
		}
	}
}
 
