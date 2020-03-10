// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package platform

import (
	"sync"
	"time"
)

type TimeProvider struct {
	now time.Time
	mu  sync.RWMutex
}

func (np *TimeProvider) Now() time.Time {
	np.mu.RLock()
	defer np.mu.RUnlock()
	return np.now
}

func NewTimeProvider(ticker time.Duration) *TimeProvider {
	p := &TimeProvider{now: time.Now()}
	go func() {
		t := time.NewTicker(ticker)
		for now := range t.C {
			p.mu.Lock()
			p.now = now
			p.mu.Unlock()
		}
	}()
	return p
}
