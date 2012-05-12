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
    lock := new(TASLock)
    wg := new(sync.WaitGroup)
    wg.Add(procs)
    for proc := 0; proc < procs; proc++ {
        go func() {
            runtime.LockOSThread()
            for i := 0; i < b.N; i++ {
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
    lock := new(TTASLock)
    wg := new(sync.WaitGroup)
    wg.Add(procs)
    for proc := 0; proc < procs; proc++ {
        go func() {
            runtime.LockOSThread()
            for i := 0; i < b.N; i++ {
                lock.Lock()
                lock.Unlock()
            }
            wg.Done()
        }()
    }
    wg.Wait()
}
