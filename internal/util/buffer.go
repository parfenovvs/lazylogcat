package util

import (
	"sync"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

type RingBuffer struct {
	lines    []model.LogLine
	head     int
	size     int
	capacity int
	mu       sync.RWMutex
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}
	return &RingBuffer{
		lines:    make([]model.LogLine, capacity),
		head:     0,
		size:     0,
		capacity: capacity,
	}
}

func (b *RingBuffer) Append(l model.LogLine) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines[b.head] = l
	b.head = (b.head + 1) % b.capacity
	if b.size < b.capacity {
		b.size++
	}
}

func (b *RingBuffer) Recent(n int) []model.LogLine {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.size == 0 {
		return []model.LogLine{}
	}
	n = min(n, b.size)
	result := make([]model.LogLine, n)
	if b.size < b.capacity {
		copy(result, b.lines[b.head-n:b.head])
		return result
	}
	for i := range n {
		index := (b.head - n + b.capacity + i) % b.capacity
		result[i] = b.lines[index]
	}
	return result
}

func (b *RingBuffer) All() []model.LogLine {
	return b.Recent(b.size)
}

func (b *RingBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *RingBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = make([]model.LogLine, b.capacity)
	b.size = 0
	b.head = 0
}
