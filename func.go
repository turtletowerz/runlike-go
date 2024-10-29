package main

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/oci/caps"
)

type twoOf[T any] struct {
	first  T
	second T
	name   string
}

func handleCapabilities(cap twoOf[[]string]) (ret []string) {
	defaults := caps.DefaultCapabilities()

	for _, c := range cap.first {
		if !slices.Contains(defaults, "CAP_"+c) {
			ret = append(ret, "--cap-add=CAP_"+c)
		}
	}

	for _, c := range cap.second {
		if slices.Contains(defaults, "CAP_"+c) {
			ret = append(ret, "--cap-drop=CAP_"+c)
		}
	}
	return
}

func handleRestart(r container.RestartPolicy) []string {
	if r.IsNone() {
		return nil
	}

	restartStr := string(r.Name)
	if r.IsOnFailure() && r.MaximumRetryCount > 0 {
		restartStr += ":" + strconv.Itoa(r.MaximumRetryCount)
	}
	return []string{"--restart=" + restartStr}
}

func handleIsolation(iso container.Isolation) []string {
	if !iso.IsDefault() {
		return []string{"--isolation=" + string(iso)}
	}
	return nil
}

// TODO: May need more work: https://docs.docker.com/reference/cli/docker/container/run/#network
func handleNetworkMode(n container.NetworkMode) []string {
	if !(n.IsDefault() || n.IsBridge()) {
		return []string{"--network=" + string(n)}
	}
	return nil
}

// NOTE: Links are deprecated, per https://docs.docker.com/engine/network/links/
func handleLinks(links []string) (ret []string) {
	for _, l := range links {
		src, dst, _ := strings.Cut(l, ":")
		src = strings.TrimPrefix(src, "/")
		dst = strings.TrimPrefix(dst, "/")

		if src != dst {
			src += ":" + dst
		}
		ret = append(ret, "--link "+src)
	}
	return
}

func handleDevices(devices []container.DeviceMapping) (ret []string) {
	for _, device := range devices {
		deviceStr := device.PathOnHost + ":" + device.PathInContainer
		if device.CgroupPermissions != "rwm" {
			deviceStr += ":" + device.CgroupPermissions
		}
		ret = append(ret, "--device "+deviceStr)
	}
	return
}

func handleLabels(l twoOf[map[string]string]) (ret []string) {
	for k, v := range l.first {
		if iv, ok := l.second[k]; !ok || v != iv {
			ret = append(ret, l.name+"'"+k+"="+v+"'")
		}
	}
	return
}

func handleHealthcheck(health *container.HealthConfig) (ret []string) {
	if health == nil {
		return
	}

	if len(health.Test) > 0 {
		if health.Test[0] == "NONE" {
			ret = append(ret, "--no-healthcheck")
			return
		}
		ret = append(ret, "--health-cmd="+strconv.Quote(strings.Join(health.Test[1:], " "))) // TODO: This is probably not right
	}

	var opts = map[string]time.Duration{
		"--health-interval=":       health.Interval,
		"--health-retries=":        time.Duration(health.Retries),
		"--health-timeout=":        health.Timeout,
		"--health-start-interval=": health.StartInterval,
		"--health-start-period=":   health.StartPeriod,
	}

	for name, val := range opts {
		if val != 0 {
			ret = append(ret, name+val.String())
		}
	}

	return
}

func handlePorts(ctdata *types.ContainerJSON) (ret []string) {
	if ctdata.HostConfig.PublishAllPorts {
		ret = append(ret, "-P")
		return
	}

	ports := ctdata.HostConfig.PortBindings
	if ports == nil {
		return
	}

	for ctport, bindings := range ports {
		protocol := ""
		if ctport.Proto() == "udp" { // TODO: "sctp" is listed as an option, should that be included?
			protocol = "/udp"
		}

		for _, b := range bindings {
			host := ""
			if !(b.HostIP == "0.0.0.0" || b.HostIP == "" || b.HostIP == "::") {
				host = b.HostIP + ":"
			}

			// If no host port mapping is defined then we know to expose it
			if b.HostPort == "" {
				ret = append(ret, "--expose "+ctport.Port())
				break
			} else if b.HostPort != "0" {
				host += b.HostPort + ":"
			}

			ret = append(ret, "-p "+host+ctport.Port()+protocol)
		}
	}

	return
}

// https://github.com/moby/moby/blob/27.x/integration/internal/container/ops.go#L138
func handleTmpFS(tmpfs map[string]string) (ret []string) {
	for k, v := range tmpfs {
		ret = append(ret, "--tmpfs="+k+":"+v)
	}
	return
}

func handleCommand(cmds twoOf[[]string]) []string {
	if len(cmds.first) == 0 {
		return nil
	}

	if slices.Compare(cmds.first, cmds.second) == 0 {
		return nil
	}

	// TODO: Escape quotes
	return []string{cmds.name + strconv.Quote(strings.Join(cmds.first, " "))}
}
