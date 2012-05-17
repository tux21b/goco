// Copyright (c) 2012 by Christoph Hack <christoph@tux21b.org>
// All rights reserved. Distributed under the Simplified BSD License.

// TODO(tux21b): write proper test cases and benchmarks

package goco

import (
    "sync"
    "testing"
)

func TestList(t *testing.T) {
    list := &List{}
    list.Add("foo")
    list.Add("bar")
    list.Add("hah")
    list.Add("foo")
    list.Remove("foo")
    list.Add("foo")
    list.Remove("bar")
}

func BenchmarkList(b *testing.B) {
    list := &List{}
    wg := &sync.WaitGroup{}

    inserter := func(values ...string) {
        for i := 0; i < b.N; i++ {
            for _, v := range values {
                list.Add(v)
            }
        }
        wg.Done()
    }

    getter := func(values ...string) {
        for i := 0; i < b.N; i++ {
            for _, v := range values {
                list.Contains(v)
            }
        }
        wg.Done()
    }

    remover := func(values ...string) {
        for i := 0; i < b.N; i++ {
            for _, v := range values {
                list.Remove(v)
            }
        }
        wg.Done()
    }

    wg.Add(7)
    go inserter("foo", "bar", "blub", "bla")
    go inserter("blub")
    go remover("foo", "bla", "quak")
    go getter("foo", "bar", "boo")
    go getter("foo", "quak")
    go getter("foo", "bar", "boo")
    go getter("foo", "quak")

    wg.Wait()
}
