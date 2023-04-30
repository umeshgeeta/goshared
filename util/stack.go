/*
 * Copyright (c) 2020. Neosemantix, Inc.
 * Author: Umesh Patil
 */

// A simple generic stack implementation.

package util

import (
	"github.com/pkg/errors"
	"sync"
)

type Stack[T any] struct { // as in:		s := Stack[int]{}
	stackSlice []T
	mux        sync.Mutex
}

func (s *Stack[T]) Empty() bool {
	if len(s.stackSlice) == 0 {
		return true
	}
	return false
}

func (s *Stack[T]) Peek() (T, error) {
	if !s.Empty() {
		return s.stackSlice[s.Size()-1], nil
	}
	return *new(T), errors.New("stack is empty")
}

func (s *Stack[T]) Pop() (T, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.Empty() {
		result := s.stackSlice[s.Size()-1]
		s.stackSlice = s.stackSlice[:s.Size()-1]
		return result, nil
	}
	return *new(T), errors.New("stack is empty")
}

func (s *Stack[T]) Push(top T) {
	s.mux.Lock()
	s.stackSlice = append(s.stackSlice, top)
	s.mux.Unlock()
}

func (s *Stack[T]) Size() int {
	return len(s.stackSlice)
}
