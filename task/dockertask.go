package task

import (
	"context"
	"os/exec"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"

	//	"github.com/moby/moby/api/types/image"
	"github.com/docker/docker/pkg/archive"
	"github.com/moby/moby/client"
	//	"github.com/moby/moby/pkg/stdcopy"
	"io"
	// "github.com/moby/moby/api/types/container"
	// "github.com/moby/moby/api/types/image"
	// "github.com/moby/moby/client"
	"log"
	"math"
	"os"

	"github.com/docker/docker/pkg/stdcopy"
	// "github.com/vishvananda/netlink"
	// "github.com/vishvananda/netns"
	"strconv"
)

type Docker struct {
	Client *client.Client
	Config Config
}

type DockerResult struct {
	Netnsid     string
	Pid         int
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run(cmd []string) DockerResult {
	ctx := context.Background()

	var dockerfilepath string

	switch d.Config.Image {
	case "debian-ovs":
		dockerfilepath = "Dockerfiles/OVS/Dockerfile"
	default:
		dockerfilepath = "Dockerfiles/DEFAULT/Dockerfile"
	}

	filectx, _ := archive.TarWithOptions(dockerfilepath, &archive.TarOptions{})
	reader, err := d.Client.ImageBuild(
		ctx, filectx, client.ImageBuildOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader.Body)

	rp := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
		Cmd:          cmd,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
		NetworkMode:     "none",
		Privileged:      true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container using image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})

	if err != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	// d.Config.Runtime.ContainerID = resp.ID

	out, err := d.Client.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	cinfo, _ := d.Client.ContainerInspect(ctx, resp.ID)

	netnsid := uuid.NewString()
	exec.Command("sudo", "ip", "netns", "attach", netnsid, strconv.Itoa(cinfo.State.Pid)).Output()

	return DockerResult{Netnsid: netnsid, Pid: cinfo.State.Pid, ContainerId: resp.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container %v", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, client.ContainerStopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if err != nil {
		log.Printf("Error removing container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}
