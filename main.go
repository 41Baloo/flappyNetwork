package main

import (
	"encoding/gob"
	game "flappyNetwork/flap"
	"flappyNetwork/network"
	"flappyNetwork/util"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetTPS(-1)
	gob.Register(&network.NeuralNetwork{})
	rand.Seed(time.Now().UnixNano())

	var trainedNetwork *network.NeuralNetwork

	if len(os.Args) == 1 {
		trainedNetwork = trainFlap(10000, 100, 2, 1)
	} else {
		var loadedErr error
		trainedNetwork, loadedErr = network.LoadNeuralNetwork(os.Args[1])
		if loadedErr != nil {
			panic(loadedErr)
		}

		fmt.Println("[ Loaded Network ]")
	}

	if len(os.Args) == 1 {
		tNow := time.Now().Unix()
		err := network.SaveNeuralNetwork("NN_"+fmt.Sprint(tNow)+".gob", trainedNetwork)
		if err != nil {
			fmt.Println("Error saving neural network:", err)
		}
		fmt.Println("Saved Network To NN_" + fmt.Sprint(tNow) + ".gob")
	}

	testFlap(trainedNetwork)

}

func testFlap(trainedNetwork *network.NeuralNetwork) {

	nGame := game.NewGame(int64(rand.Int()), 2)
	nGame.Birds[0].TestMode = true
	nGame.Birds[0].TestNetwork = trainedNetwork

	go func(g *game.Game) {
		for {
			if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
				g.WaitTime += 10
			}

			if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
				g.WaitTime -= 10
				if g.WaitTime < 0 {
					g.WaitTime = 0
				}
			}

			if ebiten.IsKeyPressed(ebiten.KeyG) {
				if g.Birds[1].Godmode {
					g.Birds[1].Godmode = false
					g.Birds[1].Color = color.Black
				} else {
					g.Birds[1].Godmode = true
					g.Birds[1].Color = color.Gray{50}
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(nGame)

	err := ebiten.RunGame(nGame)
	if err != nil {
		fmt.Println(err.Error())
	}

	for !nGame.IsGameOver {
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Score: %.10f\n", nGame.GetScore())
}

func trainFlap(populationSize int, generations int, baseMutation float64, scoreMutation float64) *network.NeuralNetwork {

	population := make([]*network.NeuralNetwork, populationSize)
	for i := range population {
		population[i] = network.NewNeuralNetwork([]int{6, 32, 32, 8, 1})
	}

	var trainedResult *network.NeuralNetwork
	semaphore := make(chan struct{}, 20)
	startTime := time.Now()
	lastBestNetworkFP := ""
	lastBestScore := 0.0
	NetworkStreak := 1
	ScoreStreak := 1

	for generation := 0; generation < generations; generation++ {
		mutationRate := (baseMutation * math.Pow(float64(NetworkStreak), 2)) + (scoreMutation * math.Pow(float64(ScoreStreak), 2))
		scores := []network.NetworkScores{}
		var wg sync.WaitGroup
		results := make(chan [3]float64, populationSize*populationSize)

		for i := 0; i < populationSize; i++ {

			wg.Add(1)
			go func(i int) {
				semaphore <- struct{}{}
				defer wg.Done()

				var prediction float64

				flapGame := game.NewGame(int64(generation), 1)

				for !flapGame.IsGameOver {
					input1, input2, input3, input4, input5, input6 := flapGame.GetInputs(flapGame.Birds[0])
					prediction := population[i].Predict([]float64{input1, input2, input3, input4, input5, input6})
					flapGame.Birds[0].JumpRate = prediction
					flapGame.Update()
					tempScore := flapGame.GetScore()
					if tempScore > 50000 {
						fmt.Printf("[+] [ "+population[i].Gene+" Is Rocking ]: %.4f. Ending Training.\n", tempScore)
						flapGame.IsGameOver = true
						generation = generations + 1
					}
				}

				score := flapGame.GetScore()

				results <- [3]float64{float64(i), score, prediction}
				<-semaphore
			}(i)

		}

		go func() {
			wg.Wait()
			close(results)
		}()

		totalScore := 0.0

		for result := range results {
			networkIndex, score, prediction := result[0], result[1], result[2]
			totalScore += score
			scores = append(scores, network.NetworkScores{
				NetworkIndex: int(networkIndex),
				Score:        score,
				Prediction:   prediction,
				Fingerprint:  population[int(networkIndex)].Gene,
			})
		}

		sort.Slice(scores, func(i, j int) bool {
			return scores[i].Score > scores[j].Score
		})

		newPopulation := make([]*network.NeuralNetwork, populationSize)

		// Elitism
		eliteCount := populationSize / 10
		for i := 0; i < eliteCount; i++ {
			newPopulation[i] = population[scores[i].NetworkIndex]
		}
		crossoverCount := (populationSize - eliteCount) / 2

		// Crossover for the next segment of the new population
		for i := 0; i < crossoverCount/2; i++ {
			parent1 := population[scores[rand.Intn(eliteCount)].NetworkIndex]
			parent2 := population[scores[rand.Intn(eliteCount)].NetworkIndex]
			newPopulation[eliteCount+i*2] = parent1.Crossover(parent2)
			newPopulation[eliteCount+i*2+1] = parent2.Crossover(parent1)
		}

		// Mutation for final segment of new population
		for i := eliteCount + crossoverCount; i < populationSize; i++ {
			bestNNIndex := scores[i-crossoverCount].NetworkIndex
			newPopulation[i] = population[bestNNIndex].CopyAndMutate(mutationRate)
		}

		trainedResult = population[scores[0].NetworkIndex]

		population = newPopulation

		if lastBestNetworkFP == scores[0].Fingerprint {
			NetworkStreak++
		} else {
			lastBestNetworkFP = scores[0].Fingerprint
			NetworkStreak = 1
		}

		scoreRounded := util.RoundToDecimal(scores[0].Score, 4)
		if lastBestScore == scoreRounded {
			ScoreStreak++
		} else {
			lastBestScore = scoreRounded
			ScoreStreak = 1
		}

		fmt.Printf("[ Generation %d / %d Completed. Best Score: %.4f, Prediction: %.4f, FP: %s; Mut: %.2f; TTL: %.4f ]\n", generation+1, generations, scores[0].Score, scores[0].Prediction, scores[0].Fingerprint, mutationRate, totalScore)
	}

	trainingTime := time.Since(startTime)
	fmt.Println("[ Trained For ] ", trainingTime)

	return trainedResult
}
