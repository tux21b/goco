// Copyright (c) 2012 by Christoph Hack <christoph@tux21b.org>
// All rights reserved. Distributed under the Simplified BSD License.

package goco

import "sync/atomic"

type Lock interface {
    Lock()
    Unlock()
}

type TASLock struct {
    state int32
}

func (l *TASLock) Lock() {
    for !atomic.CompareAndSwapInt32(&l.state, 0, 1) {
    }
}

func (l *TASLock) Unlock() {
    atomic.StoreInt32(&l.state, 0)
}

type TTASLock struct {
    state int32
}

func (l *TTASLock) Lock() {
    for {
        for atomic.LoadInt32(&l.state) != 0 {
        }

        if atomic.CompareAndSwapInt32(&l.state, 0, 1) {
            return
        }
    }
}

func (l *TTASLock) Unlock() {
    atomic.StoreInt32(&l.state, 0)
}
