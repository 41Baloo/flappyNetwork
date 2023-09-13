# flappyNetwork

A neural network learning how to play flappy bird

## How to use

You can start the compiled file without any arguments to start training a new network, that will learn how to play flappy bird.

Alternatively you can specify to play against an already trained network, by specifying its name. E.g.

`main TRAINED.gob`

### Controls

When playing the actual game, there is a white square and a black one. The white square is controlled by the best performing network or the one you loaded at the beginning. The black square is controlled by you.

- `Space` is used to make your bird jump
- `R` is used to reset the game
- `G` will make **your** bird invulnerable/revive it
- `ArrowLeft` slows down the time between frames
- `ArrowRight` speeds up the time between frames

### Debug Information

In the top right are the neural networks `inputs` and its `output` displayed. Underneath them you can find the `TPS` and `FPS`, aswell as artificially introduced time between frames (the slow motion controlled via the arrow keys)

## Neural Network

The neural network that is used is an evolutionary neural network.

### Structure

The neural network has 6 `inputs` and 1 `output`, that being a float between `0.0` and `1.0` (`true`/`false`)

Each frame/tick the neural network gets these inputs:

- `gSpeed`, the current speed of the game
- `bHeight`, the height of the bird
- `bVel`, the current (falling) speed of the bird
- `pDst`, the distance to the next pipe
- `pU`, the coordinates of the upper pipe
- `pL`, the coodinates of the lower pipe

Based on these inputs, the network returns wether or not it wants to press `jump` on the current frame. If the `output` is `> 0.5` the network jumps.

### Evolution

To start, `{population size}` networks are created with random weights and biases. 

Based on the fitness function, in this case how far the bird controlled by the network came, the network gets an amount of points. The best networks make it to the next generation without any modifications (`elitism`/"survival of the fittest"). Half of the rest gets `crossed over` with the previously mentioned best networks (procreation). The rest gets `randomly mutated` based on a mutation score

This process is repeated until the desired amount of `generations` is met or until `50.000` points are scored by a network, which signals the network has evolved enough

All of this means that the programm is able to train a network to play flappy bird without any help from the outside, simply from knowing which network performed good and which one did not

## Purpose

This doesn't really serve any purpose, except trying out a neural network i built from scratch. This code was mostly written late at night so don't judge it too much

## Disclaimer

While this might run on linux, i've only tested it on windows.