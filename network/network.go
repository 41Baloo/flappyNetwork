package network

import (
	"flappyNetwork/util"
	"math/rand"
)

type NetworkScores struct {
	NetworkIndex int
	Score        float64
	Prediction   float64
	Fingerprint  string
}

type NeuralNetwork struct {
	Weights    [][]float64
	Biases     []float64
	LayerSizes []int
	Gene       string
}

func NewNeuralNetwork(LayerSizes []int) *NeuralNetwork {
	nn := &NeuralNetwork{
		Weights:    make([][]float64, len(LayerSizes)-1),
		Biases:     make([]float64, len(LayerSizes)-1),
		LayerSizes: LayerSizes,
	}
	for i := 0; i < len(LayerSizes)-1; i++ {
		nn.Weights[i] = make([]float64, LayerSizes[i]*LayerSizes[i+1])
		nn.Biases[i] = rand.Float64()*2 - 1
	}
	nn.Gene = util.FingerprintGenerate(10)
	return nn
}

func (nn *NeuralNetwork) Predict(input []float64) float64 {
	currentInput := make([]float64, len(input))
	for i, val := range input {
		currentInput[i] = float64(val)
	}
	for layer, Weights := range nn.Weights {
		nextInput := make([]float64, nn.LayerSizes[layer+1])
		for i := 0; i < nn.LayerSizes[layer+1]; i++ {
			sum := nn.Biases[layer]
			for j := 0; j < nn.LayerSizes[layer]; j++ {
				sum += currentInput[j] * Weights[i*nn.LayerSizes[layer]+j]
			}
			nextInput[i] = util.Sigmoid(sum)
		}
		currentInput = nextInput
	}
	// Currently just a true false output
	return currentInput[0]
}

func (nn *NeuralNetwork) CopyAndMutate(mutationRate float64) *NeuralNetwork {
	child := &NeuralNetwork{
		Weights:    make([][]float64, len(nn.Weights)),
		Biases:     make([]float64, len(nn.Biases)),
		LayerSizes: nn.LayerSizes,
	}
	for layer := range nn.Weights {
		child.Weights[layer] = make([]float64, len(nn.Weights[layer]))
		for i := range nn.Weights[layer] {
			child.Weights[layer][i] = nn.Weights[layer][i] + rand.NormFloat64()*mutationRate
		}
		child.Biases[layer] = nn.Biases[layer] + rand.NormFloat64()*mutationRate
	}
	child.Gene = util.FingerprintMutate(nn.Gene)
	return child
}

func (nn *NeuralNetwork) Crossover(partner *NeuralNetwork) *NeuralNetwork {
	child := NewNeuralNetwork(nn.LayerSizes)

	for layer := range nn.Weights {
		crossoverPoint := rand.Intn(len(nn.Weights[layer]))

		child.Weights[layer] = make([]float64, len(nn.Weights[layer]))

		for i := range nn.Weights[layer] {
			if i < crossoverPoint {
				child.Weights[layer][i] = nn.Weights[layer][i]
			} else {
				child.Weights[layer][i] = partner.Weights[layer][i]
			}
		}

		if rand.Float64() < 0.5 {
			child.Biases[layer] = nn.Biases[layer]
		} else {
			child.Biases[layer] = partner.Biases[layer]
		}
	}
	child.Gene = util.FingerprintCombine(nn.Gene, partner.Gene)
	return child
}
