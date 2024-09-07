package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func optionalMulti(ctdata, imgdata iter.Seq[string], name string) (ret []string) {
outer:
	for ctv := range ctdata {
		for imv := range imgdata {
			if ctv == imv {
				continue outer
			}
		}
		ret = append(ret, "--"+name+"="+strconv.Quote(ctv))
	}
	return
}

func optional(opt, name string) string {
	if opt != "" {
		return name + opt
	}
	return ""
}

func optionalBool(opt bool, name string) string {
	if opt {
		return name
	}
	return ""
}

// TODO: Links, Devices and Logs use a set, is that necessary?
func parsePorts(info *types.ContainerJSON) (ret []string) {
	if info.HostConfig.PublishAllPorts {
		ret = append(ret, "-P")
		return
	}

	ports := info.NetworkSettings.Ports
	for k, v := range info.HostConfig.PortBindings {
		ports[k] = v
	}

	exposed := info.Config.ExposedPorts

	for ctport, bindings := range ports {
		// Skip exposed ports
		if _, ok := exposed[ctport]; ok {
			continue
		}

		protocol := ""
		if ctport.Proto() == "udp" { // TODO: "sctp" is listed as an option, should we include that?
			protocol = "/udp"
		}

		for _, b := range bindings {
			host_ip := ""
			if !(b.HostIP == "0.0.0.0" || b.HostIP == "") {
				host_ip = b.HostIP + ":"
			}

			host_port := ""
			if !(b.HostPort == "0" || b.HostPort == "") {
				host_port = b.HostPort + ":"
			}

			ret = append(ret, "-p"+host_ip+host_port+ctport.Port()+protocol)
		}
	}

	for e := range exposed {
		ret = append(ret, "--expose="+e.Port())
	}

	return
}

// Order defined per https://github.com/lavie/runlike/blob/master/runlike/inspector.py#L223
func parseFromJSON(cli *client.Client, info *types.ContainerJSON) ([]string, error) {
	imgdata, _, err := cli.ImageInspectWithRaw(context.Background(), info.Image)
	if err != nil {
		return nil, err
	}

	namesplit := strings.Split(info.Name, "/")

	flags := []string{
		"--name=" + namesplit[len(namesplit)-1],
		"--hostname=" + info.Config.Hostname,
		optional(info.Config.User, "--user="),
		optional(info.NetworkSettings.MacAddress, "--mac-address="), // TODO: Make sure this is guaranteed; info.Config.MacAddress is deprecated
		optional(string(info.HostConfig.PidMode), "--pid "),
		optional(info.HostConfig.CpusetCpus, "--cpuset-cpus="),
		optional(info.HostConfig.CpusetMems, "--cpuset-mems="),
	}

	nilSlice := slices.Values[[]string](nil)
	flags = append(flags, optionalMulti(slices.Values(info.Config.Env), slices.Values(imgdata.Config.Env), "env")...)
	flags = append(flags, optionalMulti(slices.Values(info.HostConfig.Binds), nilSlice, "volume")...)
	flags = append(flags, optionalMulti(maps.Keys(info.Config.Volumes), maps.Keys(imgdata.Config.Volumes), "volume")...)
	flags = append(flags, optionalMulti(slices.Values(info.HostConfig.VolumesFrom), nilSlice, "volumes-from")...)
	flags = append(flags, optionalMulti(slices.Values(info.HostConfig.CapAdd), nilSlice, "cap-add")...)
	flags = append(flags, optionalMulti(slices.Values(info.HostConfig.CapDrop), nilSlice, "cap-drop")...)
	flags = append(flags, optionalMulti(slices.Values(info.HostConfig.DNS), nilSlice, "dns")...)

	if networkMode := info.HostConfig.NetworkMode; !networkMode.IsDefault() {
		flags = append(flags, "--network="+string(networkMode))
	}

	flags = append(flags, optionalBool(info.HostConfig.Privileged, "--privileged"))
	flags = append(flags, optional(info.Config.WorkingDir, "--workdir="))

	// Ports
	flags = append(flags, parsePorts(info)...)

	// NOTE: Links are deprecated, per https://docs.docker.com/engine/network/links/
	for _, l := range info.HostConfig.Links {
		split := strings.Split(l, ":")
		src := strings.TrimPrefix(split[0], "/")
		dst := strings.TrimPrefix(split[1], "/")

		if src != dst {
			flags = append(flags, "--link "+src+":"+dst)
		} else {
			flags = append(flags, "--link "+src)
		}
	}

	// Restart Policy
	if restart := info.HostConfig.RestartPolicy; !restart.IsNone() {
		restartStr := string(restart.Name)
		if restart.IsOnFailure() && restart.MaximumRetryCount > 0 {
			restartStr += ":" + strconv.Itoa(restart.MaximumRetryCount)
		}
		flags = append(flags, "--restart="+restartStr)
	}

	// Devices
	for _, device := range info.HostConfig.Devices {
		deviceStr := device.PathOnHost + ":" + device.PathInContainer
		if device.CgroupPermissions != "rwm" {
			deviceStr += ":" + device.CgroupPermissions
		}
		flags = append(flags, "--device "+deviceStr)
	}

	// Labels
	ctlabels := info.Config.Labels
	imglabels := imgdata.Config.Labels
	for k, v := range ctlabels {
		if iv, ok := imglabels[k]; !ok || v != iv {
			flags = append(flags, "--label='"+k+"="+v+"'")
		}
	}

	// Logging
	if logType := info.HostConfig.LogConfig.Type; logType != "json-file" {
		flags = append(flags, "--log-driver="+logType)
	}

	for k, v := range info.HostConfig.LogConfig.Config {
		flags = append(flags, "--log-opt "+k+"="+v)
	}

	// Hosts and Runtime
	for _, h := range info.HostConfig.ExtraHosts {
		flags = append(flags, "--add-host "+h)
	}

	flags = append(flags, optional(info.HostConfig.Runtime, "--runtime="))

	// Resources
	if mem := int(info.HostConfig.Memory); mem != 0 {
		flags = append(flags, "--memory=\""+strconv.Itoa(mem)+"\"")
	}

	if reserved := int(info.HostConfig.MemoryReservation); reserved != 0 {
		flags = append(flags, "--memory-reservation=\""+strconv.Itoa(reserved)+"\"")
	}

	// Misc additional flags
	flags = append(flags, optionalBool(!info.Config.AttachStdout, "-d"))
	flags = append(flags, optionalBool(info.Config.Tty, "-t"))
	flags = append(flags, optionalBool(info.HostConfig.AutoRemove, "--rm"))
	flags = append(flags, info.Config.Image)

	// TODO: Should commands be quoted?
	flags = append(flags, strings.Join(info.Config.Cmd, " "))

	return flags, nil
}

func parseFromName(cli *client.Client, name string) ([]string, error) {
	ctdata, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return parseFromJSON(cli, &ctdata)
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
		sep = "\\\n\t"
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
