package core

import "sync/atomic"

type Clicker struct {
	enabled atomic.Bool
	cps     atomic.Int64
}

func NewClicker(initialCPS int) *Clicker {
	c := &Clicker{}
	c.cps.Store(int64(initialCPS))
	return c
}

func (c *Clicker) Toggle() bool {
	for {
		current := c.enabled.Load()
		if c.enabled.CompareAndSwap(current, !current) {
			return !current
		}
	}
}

func (c *Clicker) Enabled() bool {
	return c.enabled.Load()
}

func (c *Clicker) SetEnabled(value bool) bool {
	for {
		current := c.enabled.Load()
		if current == value {
			return current
		}
		if c.enabled.CompareAndSwap(current, value) {
			return value
		}
	}
}

func (c *Clicker) SetCPS(value int) int {
	if value <= 0 {
		return int(c.cps.Load())
	}

	c.cps.Store(int64(value))
	return value
}

func (c *Clicker) CPS() int {
	return int(c.cps.Load())
}
