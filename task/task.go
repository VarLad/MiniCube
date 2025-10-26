package task

import (
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"

	//	"github.com/moby/moby/api/types/image"
	//	"github.com/moby/moby/pkg/stdcopy"
	// "github.com/moby/moby/api/types/container"
	// "github.com/moby/moby/api/types/image"
	// "github.com/moby/moby/client"
	"time"
	// "github.com/vishvananda/netlink"
	// "github.com/vishvananda/netns"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

type Task struct {
	ID            uuid.UUID
	ContainerID   string
	Name          string
	State         State
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}
