# runlike-go
[runlike](https://github.com/lavie/runlike) in Go, providing a smaller Docker image.

## Motivation
This is a project meant mostly to get me acquainted with Github Actions. I chose to recreate this library because
I found the original project's Docker image to be 400MB+, which I thought was excessive for such a simple utility.
I wanted to recreate one that would (ideally) be around 10MB, trimming out the actual build process which is already
done in a separate Github Action.

## Usage
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/turtletowerz/runlike-go YOUR-CONTAINER`

## Flag Support
This list contains all flags offered as of **Docker v27.2** / **API v1.47**

| Flag | Supported | Notes |
| ---- | :---------: | ----- |
|`--add-host` | ✔ ||
|`--annotation` | ✔ ||
|`-a, --attach` | ✔ | `-a` with no arguments is the same as `--attach stdin` |
|`--blkio-weight` | ✔ ||
|`--blkio-weight-device` | ✔ ||
|`--cap-add` | ✔ ||
|`--cap-drop` | ✔ ||
|`--cgroup-parent` | ✔ ||
|`--cgroupns` | ✖ | Unsure of what the default value should be (varies with Docker version) |
|`--cidfile` | ✔ ||
|`--cpu-count` | ✖ | Windows only? |
|`--cpu-percent` | ✖ | Windows only? |
|`--cpu-period` | ✔ ||
|`--cpu-quota` | ✔ ||
|`--cpu-rt-period` | ✔ ||
|`--cpu-rt-runtime` | ✔ ||
|`-c, --cpu-shares` | ✔ ||
|`--cpus` | ✔ ||
|`--cpuset-cpus` | ✔ ||
|`--cpuset-mems` | ✔ ||
|`-d, --detach` | ✖ | Not stored in the container's config |
|`--detach-keys` | ✖ | Not stored in the container's config |
|`--device` | ✔ ||
|`--device-cgroup-rule` | ✔ ||
|`--device-read-bps` | ✔ ||
|`--device-read-iops` | ✔ ||
|`--device-write-bps` | ✔ ||
|`--device-write-iops` | ✔ ||
|`--disable-content-trust` | ✖ | Not stored in the container's config |
|`--dns` | ✔ ||
|`--dns-option` | ✔ ||
|`--dns-search` | ✔ ||
|`--domainname` | ✔ ||
|`--entrypoint` | ✔ ||
|`-e, --env` | ✔ ||
|`--env-file` | ✖ | Merged with `--env` in the container config |
|`--expose` | ✔ ||
|`--gpus` | ✖ | Not sure how to separate from devices |
|`--group-add` | ✔ ||
|`--health-cmd` | ✔ ||
|`--health-interval` | ✔ ||
|`--health-retries` | ✔ ||
|`--health-start-interval` | ✔ ||
|`--health-start-period` | ✔ ||
|`--health-timeout` | ✔ ||
|`--help`|||
|`-h, --hostname` | ✔ ||
|`--init` | ✔ ||
|`-i, --interactive` | ✖ | Not stored in the container's config |
|`--io-maxbandwidth` | ✖ | Windows only? |
|`--io-maxiops` | ✖ | Windows only? |
|`--ip` | ✖ | How to differentiate from default? |
|`--ip6` | ✖ | How to differentiate from default? |
|`--ipc` | ✖ | Unsure of what the default value should be (varies with Docker version) |
|`--isolation` | ✔ ||
|`--kernel-memory` | ✔ ||
|`-l, --label` | ✔ ||
|`--label-file` | ✖ | Merged with `--label` in the container config |
|`--link` | ✔ | Marked as [Deprecated](https://docs.docker.com/engine/network/links/) |
|`--link-local-ip` | ✖ | [Deprecated](https://docs.docker.com/reference/api/engine/version-history/#v144-api-changes) and ignored by Docker |
|`--log-driver` | ✔ ||
|`--log-opt` | ✔ ||
|`--mac-address` | ✖ | How to differentiate from default? |
|`-m, --memory` | ✔ ||
|`--memory-reservation` | ✔ ||
|`--memory-swap` | ✔ ||
|`--memory-swappiness` | ✔ ||
|`--mount` | ✖ | WIP |
|`--name` | ✔ ||
|`--network` | ✖ | WIP |
|`--network-alias` | ✖ | WIP |
|`--no-healthcheck` | ✔ ||
|`--oom-kill-disable` | ✔ ||
|`--oom-score-adj` | ✔ ||
|`--pid` | ✔ ||
|`--pids-limit` | ✔ ||
|`--platform` | ✖ | Not stored in the container's config |
|`--privileged` | ✔ ||
|`-p, --publish` | ✔ ||
|`-P, --publish-all` | ✔ ||
|`--pull` | ✖ | Not stored in the container's config |
|`-q, --quiet` | ✖ | Not stored in the container's config |
|`--read-only` | ✔ ||
|`--restart` | ✔ ||
|`--rm` | ✔ ||
|`--runtime` | ✔ ||
|`--security-opt` | ✔ ||
|`--shm-size` | ✖ | WIP, Cannot find the default value |
|`--sig-proxy` | ✖ | Not stored in the container's config |
|`--stop-signal` | ✔ ||
|`--stop-timeout` | ✔ ||
|`--storage-opt` | ✔ ||
|`--sysctl` | ✔ ||
|`--tmpfs` | ✔ ||
|`-t, --tty` | ✔ ||
|`--ulimit` | ✔ ||
|`-u, --user` | ✔ ||
|`--userns` | ✔ ||
|`--uts` | ✔ ||
|`-v, --volume` | ✔ ||
|`--volume-driver` | ✔ ||
|`--volumes-from` | ✔ ||
|`-w, --workdir` | ✔ ||
