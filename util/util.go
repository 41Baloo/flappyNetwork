package util

import (
	"math"
	"math/rand"
)

func Sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func Softmax(x []float64) []float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	exps := make([]float64, len(x))
	sum := 0.0

	for i, v := range x {
		exps[i] = math.Exp(v - max) // Subtract max for numerical stability
		sum += exps[i]
	}

	for i := range exps {
		exps[i] /= sum
	}
	return exps
}

func SliceContains(s []int, val int) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}

	return false
}

func SimilarityScore(target float64, actual float64, maximumReward float64, lowestVal float64, highestVal float64) float64 {
	// Handle edge case when lowestVal is equal to highestVal
	if lowestVal == highestVal {
		if actual == target {
			return maximumReward
		}
		return 0
	}

	// Find out which boundary value is closer to the target
	if math.Abs(target-lowestVal) < math.Abs(target-highestVal) {
		normalizedDiff := (actual - lowestVal) / (highestVal - lowestVal)
		return math.Max(0, maximumReward-(normalizedDiff*maximumReward))
	} else {
		normalizedDiff := (highestVal - actual) / (highestVal - lowestVal)
		return math.Max(0, maximumReward-(normalizedDiff*maximumReward))
	}
}
func XORAnswer(i int, j int) float64 {
	if i == 0 && j == 0 {
		return 0
	} else if i == 0 && j == 1 {
		return 1
	} else if i == 1 && j == 1 {
		return 0
	} else {
		return 1
	}
}

func BoolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func Quantize(value float64) byte {
	// Map the value from [-1, 1] to [0, 255]
	return byte((value + 1) * 127.5)
}

// Mutation rate (5% chance for each bit to mutate)
const mutationRate = 0.05

func FingerprintCombine(fp1, fp2 string) string {
	if len(fp1) != len(fp2) {
		return ""
	}

	result := make([]byte, len(fp1))

	for i := 0; i < len(fp1); i++ {
		if i%2 == 0 {
			result[i] = fp1[i]
		} else {
			result[i] = fp2[i]
		}
	}

	return string(result)
}

// Introduce random mutations based on the mutation rate.
// Ensure the mutated fingerprint is different from the original.
func FingerprintMutate(fp string) string {
	mutated := make([]byte, len(fp))
	copy(mutated, []byte(fp))

	for {
		changed := false
		for i := 0; i < len(fp); i++ {
			if rand.Float64() < mutationRate {
				// Flip the bit
				if mutated[i] == '0' {
					mutated[i] = '1'
				} else {
					mutated[i] = '0'
				}
				changed = true
			}
		}
		if changed && string(mutated) != fp {
			break
		}
	}

	return string(mutated)
}

// Generate a random fingerprint of the given length
func FingerprintGenerate(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		if rand.Intn(2) == 0 {
			result[i] = '0'
		} else {
			result[i] = '1'
		}
	}
	return string(result)
}

func RoundToDecimal(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(f*shift) / shift
}
