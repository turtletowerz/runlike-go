# runlike-go
[runlike](https://github.com/lavie/runlike) in Go, providing a smaller Docker image.


### Motivation
This is a project meant mostly to get me acquainted with Github Actions. I chose to recreate this library because
I found the original project's Docker image to be 300MB, which I thought was extremely large for such a simple library.
I wanted to recreate one that would (ideally) be around 10MB, trimming out the actual build process which is already
done in a separate Github Action.