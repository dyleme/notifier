package sequences

import (
	"math/rand/v2"
	"sync"
)

func New[T any](first T, next func(t T) T) Sequence[T] {
	return Sequence[T]{
		x:    first,
		next: next,
		mx:   sync.Mutex{},
	}
}

type Sequence[T any] struct {
	x    T
	next func(t T) T
	mx   sync.Mutex
}

func (s *Sequence[T]) Next() T {
	s.mx.Lock()
	defer s.mx.Unlock()
	curr := s.x
	s.x = s.next(s.x)

	return curr
}

func (s *Sequence[T]) Generate(amount int) []T {
	ts := make([]T, 0, amount)
	for i := 0; i < amount; i++ {
		ts = append(ts, s.Next())
	}

	return ts
}

func NewInt() Sequence[int] {
	return New(0, func(i int) int {
		i++

		return i
	})
}

func NewRandInt() Sequence[int] {
	return New(rand.IntN(1<<16), func(i int) int {
		i++

		return i
	})
}
