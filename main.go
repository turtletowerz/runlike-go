package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func parseFromJSON(cli *client.Client, ct *types.ContainerJSON) ([]string, error) {
	imgdata, _, err := cli.ImageInspectWithRaw(context.Background(), ct.Image)
	if err != nil {
		return nil, errors.Wrap(err, "getting container image")
	}

	// TODO: --platform, --disable-content-trust (untrusted), --pull, --quiet, --detach, --sig-proxy, and --detach-keys are not stored in the container config so we cannot inspect them, how should this be solved?

	namesplit := strings.Split(ct.Name, "/")
	flags := []string{"docker run", "--name=" + namesplit[len(namesplit)-1]}

	options := []option{
		opt[bool]{ct.Config.OpenStdin, false, "-i"},
		opt[bool]{ct.Config.Tty, false, "-t"},
		opt[bool]{ct.HostConfig.AutoRemove, false, "--rm"},
		opt[bool]{ct.HostConfig.Privileged, false, "--privileged"},
		opt[string]{ct.Config.User, "", "--user="},
		optFunc[container.RestartPolicy]{ct.HostConfig.RestartPolicy, handleRestart},
		optSlice[string]{ct.Config.Env, imgdata.Config.Env, "--env="},

		// Volumes
		optSlice[string]{ct.HostConfig.Binds, nil, "-v "},
		optSlice[string]{ct.HostConfig.VolumesFrom, nil, "--volumes-from="},
		opt[string]{ct.HostConfig.VolumeDriver, "", "--volume-driver="},

		// Misc popular options
		opt[string]{ct.Config.WorkingDir, imgdata.Config.WorkingDir, "--workdir="},
		opt[string]{ct.HostConfig.LogConfig.Type, "json-file", "--log-driver="},
		optMap{ct.HostConfig.LogConfig.Config, "--log-opt "},
		optFunc[*labels]{&labels{ct.Config.Labels, imgdata.Config.Labels}, handleLabels},
		optFunc[*capabilities]{&capabilities{ct.HostConfig.CapAdd, ct.HostConfig.CapDrop}, handleCapabilities},
		opt[bool]{ct.HostConfig.ReadonlyRootfs, false, "--read-only"},

		// Lesser used options
		optFunc[[]container.DeviceMapping]{ct.HostConfig.Devices, handleDevices},
		optFunc[[]string]{ct.HostConfig.Links, handleLinks},
		opt[string]{ct.HostConfig.Runtime, "runc", "--runtime="},
		opt[container.PidMode]{ct.HostConfig.PidMode, "", "--pid "},
		optPtr[int64]{ct.HostConfig.PidsLimit, -1, "--pids-limit="},

		// Networking stuff
		// TODO: Put hostname, MAC address and other network settings behind an optional network flag?
		optFunc[*types.ContainerJSON]{ct, handlePorts},
		//opt[string]{ct.Config.Hostname, "", "--hostname="}
		//opt[string]{ct.NetworkSettings.MacAddress, "", "--mac-address="},
		optFunc[container.NetworkMode]{ct.HostConfig.NetworkMode, handleNetworkMode},
		optSlice[string]{ct.HostConfig.ExtraHosts, nil, "--add-host "},
		optSlice[string]{ct.HostConfig.DNS, nil, "--dns="},
		optSlice[string]{ct.HostConfig.DNSOptions, nil, "--dns-option="},
		optSlice[string]{ct.HostConfig.DNSSearch, nil, "--dns-search="},
		opt[string]{ct.Config.Domainname, "", "--domainname="},

		optFunc[*container.HealthConfig]{ct.Config.Healthcheck, handleHealthcheck},
		opt[time.Duration]{ct.Config.Healthcheck.Interval, 0, "--health-interval="},
		opt[int]{ct.Config.Healthcheck.Retries, 0, "--health-retries="},
		opt[time.Duration]{ct.Config.Healthcheck.Timeout, 0, "--health-timeout="},
		opt[time.Duration]{ct.Config.Healthcheck.StartInterval, 0, "--health-start-interval="},
		opt[time.Duration]{ct.Config.Healthcheck.StartPeriod, 0, "--health-start-period="},

		//////////////////////////////////////////////////////////////////////////
		// Less common options that can go at the end if needed
		opt[bool]{ct.Config.AttachStdin, false, "--attach stdin"},
		opt[bool]{ct.Config.AttachStdout, false, "-attach stdout"},
		opt[bool]{ct.Config.AttachStderr, false, "-attach stderr"},
		opt[string]{ct.HostConfig.ContainerIDFile, "", "--cidfile "},

		// TODO: Windows only?
		// opt[int]{ct.HostConfig.CPUCount, 0, "--cpu-count="},
		// opt[int]{ct.HostConfig.CPUPercent, 0, "--cpu-percent="},
		// optMem{int64(ct.HostConfig.IOMaximumBandwidth)}
		// optMem{int64(ct.HostConfig.IOMaximumIOps)}

		opt[int64]{ct.HostConfig.CPUPeriod, 0, "--cpu-period="},
		opt[int64]{ct.HostConfig.CPUQuota, 0, "--cpu-quota="},
		opt[int64]{ct.HostConfig.CPURealtimePeriod, 0, "--cpu-rt-period="},
		opt[int64]{ct.HostConfig.CPURealtimeRuntime, 0, "--cpu-rt-runtime="},
		opt[int64]{ct.HostConfig.CPUShares, 0, "--cpu-shares="},
		opt[opts.NanoCPUs]{opts.NanoCPUs(ct.HostConfig.NanoCPUs), 0, "--cpus="},
		opt[string]{ct.HostConfig.CpusetCpus, "", "--cpuset-cpus="},
		opt[string]{ct.HostConfig.CpusetMems, "", "--cpuset-mems="},

		opt[opts.MemBytes]{opts.MemBytes(ct.HostConfig.Memory), 0, "--memory="},
		opt[opts.MemBytes]{opts.MemBytes(ct.HostConfig.MemoryReservation), 0, "--memory-reservation="},
		opt[opts.MemSwapBytes]{opts.MemSwapBytes(ct.HostConfig.MemorySwap), 0, "--memory-swap="},
		optPtr[int64]{ct.HostConfig.MemorySwappiness, -1, "--memory-swappiness="},
		opt[opts.MemBytes]{opts.MemBytes(ct.HostConfig.KernelMemory), 0, "--kernel-memory="},

		optPtr[bool]{ct.HostConfig.OomKillDisable, false, "--oom-kill-disable"},
		opt[int]{ct.HostConfig.OomScoreAdj, 0, "--oom-score-adj="},

		// TODO: Seems to be no way to find the default for this
		//opt[opts.MemBytes]{opts.MemBytes(ct.HostConfig.ShmSize), 0, "--shm-size="},

		optSlice[string]{ct.HostConfig.DeviceCgroupRules, nil, "--device-cgroup-rule="},
		opt[string]{ct.HostConfig.CgroupParent, "", "--cgroup-parent="},
		//opt[container.CgroupnsMode]{ct.HostConfig.CgroupnsMode, "", "--cgroupns="},
		opt[container.UsernsMode]{ct.HostConfig.UsernsMode, "", "--userns="},
		opt[container.UTSMode]{ct.HostConfig.UTSMode, "", "--uts="},
		optSlice[string]{ct.HostConfig.GroupAdd, nil, "--group-add "},

		optSlice[*container.Ulimit]{ct.HostConfig.Ulimits, nil, "--ulimit "},

		opt[uint16]{ct.HostConfig.BlkioWeight, 0, "--blkio-weight="},
		optSlice[*blkiodev.WeightDevice]{ct.HostConfig.BlkioWeightDevice, nil, "--blkio-weight-device="},
		optSlice[*blkiodev.ThrottleDevice]{ct.HostConfig.BlkioDeviceReadBps, nil, "--blkio-read-bps="},
		optSlice[*blkiodev.ThrottleDevice]{ct.HostConfig.BlkioDeviceReadIOps, nil, "--blkio-read-iops="},
		optSlice[*blkiodev.ThrottleDevice]{ct.HostConfig.BlkioDeviceWriteBps, nil, "--blkio-write-bps="},
		optSlice[*blkiodev.ThrottleDevice]{ct.HostConfig.BlkioDeviceWriteIOps, nil, "--blkio-write-iops="},

		opt[string]{ct.Config.StopSignal, "", "--stop-signal="},
		optPtr[int]{ct.Config.StopTimeout, -1, "--stop-timeout="},

		optSlice[string]{ct.HostConfig.SecurityOpt, nil, "--security-opt="},
		optMap{ct.HostConfig.StorageOpt, "--storage-opt "},
		optMap{ct.HostConfig.Sysctls, "--sysctl "},

		optFunc[container.Isolation]{ct.HostConfig.Isolation, handleIsolation},

		//opt[container.IpcMode]{ct.HostConfig.IpcMode, "", "--ipc="},

		optMap{ct.HostConfig.Annotations, "--annotation "},

		//////////////////////////////////////////////////////////////////////////
	}

	for _, v := range options {
		if vals := v.Values(); v != nil {
			flags = append(flags, vals...)
		}
	}

	flags = append(flags, ct.Config.Image)

	if cmd := ct.Config.Cmd; cmd != nil {
		flags = append(flags, strings.Join(cmd, " "))
	}

	return flags, nil
}

func parseFromName(cli *client.Client, name string) ([]string, error) {
	ct, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return parseFromJSON(cli, &ct)
}

func _main(ctx *cli.Context) error {
	if ctx.NArg() == 0 && !ctx.Bool("stdin") {
		return errors.New("no arguemnts, provide [container] or --stdin")
	}

	ctcli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer ctcli.Close()

	var flags []string
	if ctx.Bool("stdin") {
		var data []types.ContainerJSON
		if err = json.NewDecoder(os.Stdin).Decode(&data); err != nil {
			return errors.Wrap(err, "decoding container info from STDIN")
		}

		// TODO: Maybe allow this?
		if len(data) > 1 {
			return errors.New("only 1 container can be inspected at a time")
		}

		flags, err = parseFromJSON(ctcli, &data[0])
	} else {
		flags, err = parseFromName(ctcli, ctx.Args().First())
	}

	if err != nil {
		return err
	}

	// Remove the first flag (name) if it's not wanted
	if ctx.Bool("no-name") {
		flags = flags[1:]
	}

	sep := " "
	if ctx.Bool("pretty") {
		sep = " \\\n\t"
	}

	fmt.Println(strings.Join(flags, sep))
	return nil
}

func main() {
	app := cli.App{
		Name:    "runlike",
		Version: "1.0.0",
		Usage:   "Prints the command line version of a Docker container",
		Action:  _main,

		Args:            true,
		ArgsUsage:       "[container]",
		HideHelpCommand: true,

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-name",
				Usage: "Do not include container name in output",
			},
			&cli.BoolFlag{
				Name:    "pretty",
				Aliases: []string{"p"},
				Usage:   "Pretty-print the command",
			},
			&cli.BoolFlag{
				Name:    "stdin",
				Aliases: []string{"s"},
				Usage:   "Accept input from STDIN",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln("Error: " + err.Error())
	}
}
