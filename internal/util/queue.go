package util

import "errors"

// Queue structure for shifting out pixels to the display or enqueue CPU
// micro-operations.
// We make it a fixed size that works for CPU operations as well as PPU
// pixels.
type Queue[T any] struct {
	fifo [16]T // Array of values in the Queue.
	out  int   // Current index of the tail (output) of the Queue.
	in   int   // Current index of the head (input) of the Queue.
	len  int   // Current length of the Queue.
}

// Pre-defined errors to only instantiate them once.
var errFIFOOverflow = errors.New("queue buffer overflow")
var errFIFOUnderrun = errors.New("queue buffer underrun")
var errFIFOAccessOutOfBounds = errors.New("peek into queue out of bounds")

// Push an item to the Queue.
func (q *Queue[T]) Push(item T) {
	if q.len == len(q.fifo) {
		panic(errFIFOOverflow)
	}
	q.fifo[q.in] = item
	q.in = (q.in + 1) % len(q.fifo)
	q.len++
}

// Pop an item out of the Queue.
func (q *Queue[T]) Pop() (item T, err error) {
	if q.len == 0 {
		return *new(T), errFIFOUnderrun
	}
	item = q.fifo[q.out]
	q.out = (q.out + 1) % len(q.fifo)
	q.len--
	return item, nil
}

func (q *Queue[T]) Peek(index int) (item T, err error) {
	if index > q.len-1 {
		return *new(T), errFIFOAccessOutOfBounds
	}
	return q.fifo[(q.out+index)%16], nil
}

func (q *Queue[T]) Set(index int, item T) error {
	if index > q.len-1 {
		return errFIFOAccessOutOfBounds
	}
	q.fifo[(q.out+index)%16] = item
	return nil
}

// Size returns the current number of items in the Queue.
func (q *Queue[T]) Size() int {
	return q.len
}

// Clear resets internal indexes, effectively clearing out the Queue.
func (q *Queue[T]) Clear() {
	q.in, q.out, q.len = 0, 0, 0
}
