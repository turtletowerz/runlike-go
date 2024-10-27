# runlike-go
[runlike](https://github.com/lavie/runlike) in Go, providing a smaller Docker image.


### Motivation
This is a project meant mostly to get me acquainted with Github Actions. I chose to recreate this library because
I found the original project's Docker image to be 400MB+, which I thought was excessive for such a simple utility.
I wanted to recreate one that would (ideally) be around 10MB, trimming out the actual build process which is already
done in a separate Github Action.


## Usage
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/turtletowerz/runlike-go YOUR-CONTAINER`


### Current Unsupported Flags
- `--gpus`
- `--mount`
- `--ip`
- `--ip6`
- `--network-alias`
- `--cgroupns` and `--ipc` are currently disabled, as they have have one of two options set a default depending on the Docker daemon version.
- `--link-local-ip`: Deprecated and ignored by Docker so there's no reason to support.
- `--env-file` and `--label-file` don't exist in `docker inspect` settings because they are merged with `--env` and `--label`
