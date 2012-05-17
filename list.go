// Copyright (c) 2012 by Christoph Hack <christoph@tux21b.org>
// All rights reserved. Distributed under the Simplified BSD License.
//
// Note:
// The lock-free list implementation in this file is based on a lock-free
// list design which is described in the book "The Art of Multiprocessor
// Programming". The authors of the book, Maurice Herlihy and Nir Shavit have
// mentioned that they have described a variation of a lock-free list designed
// by Maged Michael which is based on some earlier works by Tim Harris.

package goco

import (
    "sync/atomic"
    "unsafe"
)

// List represents a single-linked list of unique items, which are stored in
// ascending order. All operations of the lists are at least lock-free.
type List struct {
    head *node
}

// Add adds a new element to the list, returning true if, and only if the
// element wasn't already there. This method is lock-free.
func (l *List) Add(key string) bool {
    for {
        pred, pred_m, curr, _ := l.find(key)
        if curr != nil && curr.key == key {
            return false
        }

        node := &node{key, &markAndRef{false, curr}}

        // Insert the new node after the pred node or modify the head of the
        // list if there is no predecessor.
        if pred == nil {
            if atomic.CompareAndSwapPointer(
                (*unsafe.Pointer)(unsafe.Pointer(&l.head)),
                unsafe.Pointer(curr),
                unsafe.Pointer(node)) {
                return true
            }
        } else {
            m := &markAndRef{false, node}
            if atomic.CompareAndSwapPointer(
                (*unsafe.Pointer)(unsafe.Pointer(&pred.m)),
                unsafe.Pointer(pred_m),
                unsafe.Pointer(m)) {
                return true
            }
        }

        // Another thread has modified the pred node, by either marking it as
        // deleted or by inserting another node directly after it. The other
        // thread progressed, but we need to retry our insert.
    }
    panic("not reachable")
}

// Contains returns true if, and only if the list contains an element with that
// key. This method is wait-free, so it will always return in a finite number
// of steps, independent of any contention with other threads.
func (l *List) Contains(key string) bool {
    curr := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
    for curr != nil && curr.key < key {
        curr_m := (*markAndRef)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&curr.m))))
        curr = curr_m.next
    }
    if curr != nil && curr.key == key {
        curr_m := (*markAndRef)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&curr.m))))
        return !curr_m.marked
    }
    return false
}

// Remove removes the element with the key from the list, returning true if,
// and only if, the element was there before. This method is lock-free.
func (l *List) Remove(key string) bool {
    for {
        pred, pred_m, curr, curr_m := l.find(key)

        if curr == nil || curr.key != key {
            return false
        }

        // Mark the node as deleted.
        m := &markAndRef{true, curr_m.next}
        if !atomic.CompareAndSwapPointer(
            (*unsafe.Pointer)(unsafe.Pointer(&curr.m)),
            unsafe.Pointer(curr_m),
            unsafe.Pointer(m)) {

            // Somebody has modified the current node, either by inserting
            // directly after it, or by marking it as deleted. The other thread
            // progressed, but we need to retry our action.
            continue
        }

        // Try to remove the nody physically by unlinking it from the list. We
        // can't reuse the deleted node, because other threads might still use
        // it, but the garbage collector will take care of it once all threads
        // are done. A single attempt is enough here, since the node is already
        // marked as deleted and it will be removed during the next list
        // traversal anyway.
        if pred == nil {
            atomic.CompareAndSwapPointer(
                (*unsafe.Pointer)(unsafe.Pointer(&l.head)),
                unsafe.Pointer(curr),
                unsafe.Pointer(m.next))
        } else {
            m2 := &markAndRef{false, m.next}
            atomic.CompareAndSwapPointer(
                (*unsafe.Pointer)(unsafe.Pointer(&pred.m)),
                unsafe.Pointer(pred_m),
                unsafe.Pointer(m2))
        }

        return true
    }
    panic("not reachable")
}

// find returns the nodes of either side of a specific key. It will physically
// delete all nodes marked for deletion while traversing the list.
func (l *List) find(key string) (pred *node, pred_m *markAndRef, curr *node, curr_m *markAndRef) {
retry:
    for {
        curr = (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
        for curr != nil {
            curr_m = (*markAndRef)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&curr.m))))

            if curr_m.marked {
                // curr is marked as deleted. Try to remove it physically by
                // unlinking the node from the list.

                if pred == nil {
                    if !atomic.CompareAndSwapPointer(
                        (*unsafe.Pointer)(unsafe.Pointer(&l.head)),
                        unsafe.Pointer(curr),
                        unsafe.Pointer(curr_m.next)) {

                        // Another thread has modified the head pointer of our
                        // list. The other thread progressed, but we need to
                        // restart the list traversal.
                        continue retry
                    }
                } else {
                    m := &markAndRef{false, curr_m.next}
                    if !atomic.CompareAndSwapPointer(
                        (*unsafe.Pointer)(unsafe.Pointer(&pred.m)),
                        unsafe.Pointer(pred_m),
                        unsafe.Pointer(m)) {

                        // Another thread has progressed by modifying the next
                        // pointer of our predecessor. We need to traverse the
                        // list again.
                        continue retry
                    }
                    pred_m = m
                }
                curr = curr_m.next
                continue
            }

            if curr.key >= key {
                return
            }

            pred = curr
            pred_m = curr_m
            curr = curr_m.next
        }
        return
    }
    panic("not reachable")
}

// node represents a list node for the single linked list.
type node struct {
    key string
    m   *markAndRef
}

// markAndRef stores a boolean flag indicating if the current node is marked
// as deleted and a pointer to the next node. Instances of this struct are
// immutable. We must not modify any of its attributes after the struct
// has been created, but we can atomically replace pointers to this struct.
// This allows us to to modify the markAndRef object with both attributes (the
// flag and the next pointer) in a single atomic instruction.
type markAndRef struct {
    marked bool
    next   *node
}
