# Float 1K Gravity

<br>

## What is it?
It's a basic newtonian gravity simulation written in Go using floating point numbers with 1024 bits of accuracy. It is currently single threaded because when I tried to implement threads it totally broke. Feel free to contribute.

<br>

## Why did you make it?
Learning, for the most part. I picked this in particular because I got a gravity simulation in my recommended and I was like "ooh, that looks nice" and so I made a worse version of it.

<br>

## Known innaccuracies
- Colliding bodies' end velocity is influenced by order of collision and tickrate
- Accuracy is influenced by tickrate
- Relativity is not considered

<br>

## Controls

| Key | Action |
| --- | --- |
| L | toggle paths on/off |
| I/K | increase/decrease gravity |
| Tab | toggle debug menu |

<br>

## Compilation
There are no special compilation arguments. Just run `go build -buildmode=pie float1kgravity.go` (buildmode not acutally necessary) and you're good to go.