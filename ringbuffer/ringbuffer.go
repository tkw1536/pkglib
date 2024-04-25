// Package ringbuffer provides RingBuffer.
package ringbuffer

import "strings"

// spellchecker:words ringbuffer

// RingBuffer is a buffer holding a finite number of elements.
// Adding an element beyond the buffer's capacity, overwrites the least recently added elements.
//
// RingBuffer is in general not safe for concurrent access; however it does support concurrent
// access by multiple concurrent readers.
//
// The zero ring buffer has a capacity of zero.
// New RingBuffers should either be initialized using an explicit call to [MakeRingBuffer], or extended implicitly using [Append].
type RingBuffer[T any] struct {
	// index stores the index of the first valid element of this RingBuffer.
	// elems holds all elements of the ring buffer.
	//
	// This data structure furthermore maintains the following two invariants:
	//
	// 1. 0 <= index < len(elems) || (len(elems) == 0 && index == 0)
	// 2. len(elems) < cap(elems) implies index == 0
	index int
	elems []T
}

// MakeRingBuffer creates a new RingBuffer with the given capacity.
func MakeRingBuffer[T any](cap int) *RingBuffer[T] {
	return &RingBuffer[T]{
		index: 0,
		elems: make([]T, 0, cap),
	}
}

// Cap returns the maximum number of elements the RingBuffer can hold at the current time.
func (rb *RingBuffer[T]) Cap() int {
	return cap(rb.elems)
}

// Len returns the number of elements currently inside the RingBuffer.
func (rb *RingBuffer[T]) Len() int {
	return len(rb.elems)
}

// Add adds a new element to the ring buffer without extending it's capacity.
// This operation can be performed in O(1).
//
// If the capacity would be exceeded by adding an entry to this RingBuffer,
// this method will overwrite the most recently added element.
//
// Adding to a RingBuffer with a capacity of zero discards the element in question.
func (rb *RingBuffer[T]) Add(elem T) {
	// no space to add an element
	if cap(rb.elems) == 0 {
		return
	}

	// buffer is not yet full, by invariant we can add an element
	// and this will not allocate.
	if len(rb.elems) < cap(rb.elems) {
		rb.elems = append(rb.elems, elem)
		return
	}

	// overwrite the least recently used element
	rb.elems[rb.index] = elem

	// and increase the index
	rb.index = (rb.index + 1) % len(rb.elems)
}

// Pop deletes the most recently added element from this RingBuffer
// and returns it.
// Pop is an expensive operation, and should be avoided.
//
// If this RingBuffer contains no elements, returns the zero value.
func (rb *RingBuffer[T]) Pop() T {
	var last T

	// no elements available to pop
	if len(rb.elems) == 0 {
		return last
	}

	rb.Optimize() // ensure that rb.index == 0

	last, rb.elems[len(rb.elems)-1] = rb.elems[len(rb.elems)-1], last // flip last element and zero
	rb.elems = rb.elems[:len(rb.elems)-1]                             // decrease the capacity

	return last
}

// Push appends an element to this RingBuffer, possibly extending it's capacity.
// This operation is O(N) in the worst case; it should be avoided.
func (sr *RingBuffer[T]) Push(elem T) {
	sr.Optimize()                     // ensure that rb.index == 0
	sr.elems = append(sr.elems, elem) // append the element
}

// Optimize performs internal optimization of this RingBuffer
// so that calls to the [Elems], [Push], [Pop] and [Iterate] methods
// are slightly more efficient until the next call to [Add].
//
// Optimize is an O(N) operation; it is expensive and should only be used
// in cases where the performance of those methods is critical.
//
// If the RingBuffer is already optimized, Optimize does nothing.
func (rb *RingBuffer[T]) Optimize() {
	// This method ensures that rb.index == 0.
	// If this is already the case, we don't need to do anything.
	if rb.index == 0 {
		return
	}

	// determine the last index inside the elems array
	lastIndex := len(rb.elems) - 1

	// this is only triggered when len(rb.elems) == 1
	// or len(rb.elems) == 0.
	// In either case, by invariant 1 we must have rb.index == 0.
	// So this cannot occur.
	//
	// We leave it here for the compiler to notice
	// that the lastIndex accesses below are always valid.
	// This means that it doesn't need bounds checking.
	if lastIndex <= 0 {
		panic("never reached")
	}

	// if we are on the right hand side of the array
	// it is more efficient to do right shifts.
	if rb.index > len(rb.elems)/2 {
		// using goto to avoid indenting
		// the left and right shifts
		goto right
	}

	// perform repeated left shifts until
	// the index is at 0.
	for rb.index != 0 {
		cache := rb.elems[0]
		copy(rb.elems, rb.elems[1:])
		rb.elems[lastIndex] = cache

		rb.index--
	}
	return

	// perform right shifts until the index
	// goes beyond the end of the slice
right:
	for rb.index != len(rb.elems) {
		cache := rb.elems[lastIndex]
		copy(rb.elems[1:], rb.elems)
		rb.elems[0] = cache

		rb.index++
	}
	rb.index = 0
}

// Elems returns a copy of all elements in this RingBuffer.
func (rb *RingBuffer[T]) Elems() []T {
	// create a slice of elements
	elems := make([]T, len(rb.elems))

	// copy over the beginning of the elems we have
	n := copy(elems, rb.elems[rb.index:])

	// copy over the rest of the elements.
	// note that if n == len(sr.elems), this is automatically skipped
	copy(elems[n:], rb.elems)

	return elems
}

// Iterate iterates over the elements inside this RingBuffer.
// The iteration takes place in natural order, from oldest to newest.
// if f returns false, Iterate stops further iteration.
func (rb *RingBuffer[T]) Iterate(f func(elem T, index int) bool) bool {
	for i := range rb.elems {
		index := (i + rb.index) % len(rb.elems)
		if !f(rb.elems[index], i) {
			return false
		}
	}
	return true
}

const maxInt = int(^uint(0) >> 1)

// Join is an efficient version of strings.Join working on
// the RingBuffer.
func Join(sr *RingBuffer[string], sep string) string {
	// we didn't fill the buffer yet
	// so just fallback to a strings.Join.
	if len(sr.elems) < cap(sr.elems) {
		return strings.Join(sr.elems, sep)
	}

	// inlined version of strings.Join
	// that uses offset indexes

	// simple cases: only a few versions
	switch len(sr.elems) {
	case 0:
		return ""
	case 1:
		return sr.elems[1]
	}

	// compute the size of the output buffer
	// NOTE: This is exactly copied from strings.Join
	// because we don't care about in which order we count the sizes.
	var n int
	if len(sep) > 0 {
		if len(sep) >= maxInt/(len(sr.elems)-1) {
			panic("StringRing: Join output length overflow")
		}
		n += len(sep) * (len(sr.elems) - 1)
	}
	for _, elem := range sr.elems {
		if len(elem) > maxInt-n {
			panic("StringRing: Join output length overflow")
		}
		n += len(elem)
	}

	// create a buffer and size it appropriately
	var b strings.Builder
	b.Grow(n)

	// copy over the first element
	b.WriteString(sr.elems[sr.index])

	// we need to copy over len() - 1 elements.
	for i := range len(sr.elems) - 1 {
		b.WriteString(sep)
		// Compute the actual index of the element we need to access.
		// We start at the element after index, and go one element each step.
		i = (sr.index + (i + 1))
		b.WriteString(sr.elems[i%len(sr.elems)])
	}
	return b.String()
}
