package network

import (
	"encoding/gob"
	"os"
)

func SaveNeuralNetwork(filename string, nn *NeuralNetwork) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(nn)
	return err
}

func LoadNeuralNetwork(filename string) (*NeuralNetwork, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	nn := &NeuralNetwork{}
	err = decoder.Decode(nn)
	if err != nil {
		return nil, err
	}
	return nn, nil
}
