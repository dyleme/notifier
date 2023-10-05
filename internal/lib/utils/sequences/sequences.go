package sequences

import (
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
	s.x = s.next(s.x)

	return s.x
}

type SequenceInt struct {
	x  int
	mx sync.Mutex
}

func (s *SequenceInt) Next() int {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.x++

	return s.x
}
