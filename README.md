# Float 1K Gravity

<br>

## What is it?
It's a basic newtonian gravity simulation written in Go using floating point numbers with 1024 bits of accuracy. Distance calculation & intersection detection is multithreaded, although the effectiveness of said multithreading is arguable.

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