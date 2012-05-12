// Copyright (c) 2012 by Christoph Hack <christoph@tux21b.org>
// All rights reserved. Distributed under the Simplified BSD License.

package goco

import (
    "runtime"
    "sync"
    "testing"
)

func BenchmarkTASLock(b *testing.B) {
    procs := runtime.GOMAXPROCS(-1)
    n := b.N / procs
    lock := new(TASLock)
    wg := new(sync.WaitGroup)
    wg.Add(procs)
    for proc := 0; proc < procs; proc++ {
        go func() {
            runtime.LockOSThread()
            for i := 0; i < n; i++ {
                lock.Lock()
                lock.Unlock()
            }
            wg.Done()
        }()
    }
    wg.Wait()
}

func BenchmarkTTASLock(b *testing.B) {
    procs := runtime.GOMAXPROCS(-1)
    n := b.N / procs
    lock := new(TTASLock)
    wg := new(sync.WaitGroup)
    wg.Add(procs)
    for proc := 0; proc < procs; proc++ {
        go func() {
            runtime.LockOSThread()
            for i := 0; i < n; i++ {
                lock.Lock()
                lock.Unlock()
            }
            wg.Done()
        }()
    }
    wg.Wait()
}
