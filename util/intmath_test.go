/*
 * Copyright (c) 2020. Neosemantix, Inc.
 * Author: Umesh Patil
 */

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGcd(t *testing.T) {
	assert.Equal(t, 2, Gcd(6, 8))
	assert.Equal(t, 3, Gcd(9, 12))
	assert.Equal(t, 4, Gcd(12, 16))
}

func TestGcdList(t *testing.T) {
	assert.Equal(t, 2, GcdList([]int{6, 8, 10}))
	assert.Equal(t, 3, GcdList([]int{9, 12, 15}))
	assert.Equal(t, 4, GcdList([]int{12, 16, 20}))
}

func TestLcm(t *testing.T) {
	assert.Equal(t, 24, Lcm(6, 8))
	assert.Equal(t, 36, Lcm(9, 12))
	assert.Equal(t, 48, Lcm(12, 16))
}

func TestLcmList(t *testing.T) {
	assert.Equal(t, 120, LcmList([]int{6, 8, 10}))
	assert.Equal(t, 180, LcmList([]int{9, 12, 15}))
	assert.Equal(t, 240, LcmList([]int{12, 16, 20}))
}
