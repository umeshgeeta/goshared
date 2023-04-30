/*
 * Copyright (c) 2020. Neosemantix, Inc.
 * Author: Umesh Patil
 */

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack(t *testing.T) {
	s := Stack[int]{}
	assert.True(t, s.Empty())

	s.Push(1)
	assert.Equal(t, 1, s.Size())

	val, err := s.Peek()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, s.Size())

	val, err = s.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.True(t, s.Empty())

	_, err = s.Peek()
	assert.NotNil(t, err)

	_, err = s.Pop()
	assert.NotNil(t, err)
}
