package frontend

import (
	"errors"
	"math"
)

var (
	errIntegerOverflow = errors.New("value exceeds int64 range")
	errFloat64Overflow = errors.New("value exceeds float64 range")
)

// convertNumber converts a number represented by its integer, decimal, and exponent
// parts into either an int64 or float64.
//
// It returns an int64 if the number has no decimal part and a non-negative exponent.
// Otherwise, it returns a float64.
//
// An error is returned if the final value would overflow its target type.
func convertNumber(integerPart uint64, decimalPart uint64, hasDecimalPart bool, expPart int64) (interface{}, error) {
	// --- Integer Path ---
	// A number can only be an integer if it has no fractional part and the exponent
	// is not negative (which would create a fraction).
	if !hasDecimalPart && decimalPart == 0 && expPart >= 0 {
		if integerPart > uint64(math.MaxInt64) {
			return nil, errIntegerOverflow
		}

		result := int64(integerPart)

		// TODO binary exponent
		for i := int64(0); i < expPart; i++ {
			if result > math.MaxInt64/10 {
				return nil, errIntegerOverflow
			}
			result *= 10
		}
		return result, nil
	}

	// --- Float Path ---
	// If there's a decimal part or a negative exponent, the result must be a float.
	mantissa := float64(integerPart)
	if decimalPart > 0 {
		// For example, if decimalPart is 123, the divisor becomes 1000.0 to get 0.123
		decimal := float64(decimalPart)
		numberOfDigits := math.Floor(math.Log10(decimal) + 1)
		// TODO binary exponent
		divisor := math.Pow(10, numberOfDigits)
		mantissa += decimal / divisor
	}

	// TODO binary exponent
	finalValue := mantissa * math.Pow(10, float64(expPart))

	// math.Pow returns ±Inf on overflow
	if math.IsInf(finalValue, 0) {
		return nil, errFloat64Overflow
	}

	return finalValue, nil
}
