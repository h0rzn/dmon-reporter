package store

import (
	"sync"
)

type Buffer[T any] struct {
	mut       *sync.Mutex
	data      []T
	maxLength int
}

func NewBuffer[T any](maxLength int) *Buffer[T] {
	return &Buffer[T]{
		mut:       &sync.Mutex{},
		data:      make([]T, 0),
		maxLength: maxLength,
	}
}

func (b *Buffer[T]) Push(set T) bool {
	b.mut.Lock()
	b.data = append(b.data, set)
	b.mut.Unlock()
	return b.isFull()
}

func (b *Buffer[T]) isFull() bool {
	return len(b.data) == b.maxLength
}

func (b *Buffer[T]) Drop() []T {
	b.mut.Lock()
	dropped := make([]T, len(b.data))
	_ = copy(dropped, b.data)
	b.data = []T{}
	b.mut.Unlock()
	return dropped
}
