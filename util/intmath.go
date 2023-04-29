// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

// Absint Finds absolute value of the given integer.
func Absint(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Gcd calculates the greatest common divisor using the Euclidean algorithm.
func Gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// GcdList calculates the greatest common divisor of a list of integers.
func GcdList(nums []int) int {
	result := nums[0]
	for i := 1; i < len(nums); i++ {
		result = Gcd(result, nums[i])
	}
	return result
}

// Lcm calculates the least common multiple of two integers.
func Lcm(a, b int) int {
	return a / Gcd(a, b) * b
}

// LcmList calculates the least common multiple of a list of integers.
func LcmList(nums []int) int {
	result := nums[0]
	for i := 1; i < len(nums); i++ {
		result = Lcm(result, nums[i])
	}
	return result
}
