// %%
package main

import (
	"OrchestratorGo/cube/node"
	"OrchestratorGo/cube/task"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"

	"OrchestratorGo/cube/manager"
	"OrchestratorGo/cube/parser"
	"OrchestratorGo/cube/worker"

	"github.com/moby/moby/client"
	"slices"
	// "github.com/vishvananda/netlink"
	// "github.com/vishvananda/netns"
)

/*
func createConnections() {
	ns1 := uuid.NewString()
	ns2 := uuid.NewString()

}
*/

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "alpine:latest",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run([]string{"true"})
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result
}

func createContainersFromConfig(conf *parser.NetworkConfig) ([]*task.Docker, []*task.DockerResult, []string) {
	ds := []*task.Docker{}
	results := []*task.DockerResult{}
	hostnames := []string{}

	for hostname, host := range conf.Hosts {
		if host.Type == "docker" {

			c := task.Config{
				Name:  string(hostname),
				Image: "alpine:latest",
				Env: []string{
					"HELLO=cube",
				},
			}

			dc, _ := client.NewClientWithOpts(client.FromEnv)
			d := task.Docker{
				Client: dc,
				Config: c,
			}

			result := d.Run([]string{"sleep", "infinity"})
			if result.Error != nil {
				fmt.Printf("%v\n", result.Error)
				return nil, nil, nil
			}
			fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
			ds = append(ds, &d)
			results = append(results, &result)
			hostnames = append(hostnames, d.Config.Name)
		}
	}

	return ds, results, hostnames
}

type HostIfacePair struct {
	Host  string
	Iface string
}

func createConnectionsFromNetworkConfig(conf *parser.NetworkConfig, createResults []*task.Docker, dockerResults []*task.DockerResult, hostnames []string) {
	fmt.Println("Meow")
	usedhostpair := []HostIfacePair{}
	useddestpair := []HostIfacePair{}
	for i := range hostnames {
		for interfacename, iface := range conf.Hosts[hostnames[i]].Interfaces {
			reversepair := conf.Hosts[iface.DstNode].Interfaces[iface.DstIface]
			if slices.Contains(usedhostpair, HostIfacePair{iface.DstNode, iface.DstIface}) && slices.Contains(useddestpair, HostIfacePair{hostnames[i], interfacename}) {
				if reversepair.DstNode == string(hostnames[i]) && reversepair.DstIface == string(interfacename) {
					useddestpair = append(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface})
					usedhostpair = append(usedhostpair, HostIfacePair{hostnames[i], string(interfacename)})
					continue
				} else {
					panic("Mismatch with destination")
				}
			}
			if slices.Contains(usedhostpair, HostIfacePair{hostnames[i], string(interfacename)}) {
				panic("The same interface name or host name has been likely declared more than once.")
			} else {
				if slices.Contains(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface}) {
					panic("This interface is already in use.")
				} else {
					if reversepair.DstNode == string(hostnames[i]) && reversepair.DstIface == string(interfacename) {
						useddestpair = append(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface})
						usedhostpair = append(usedhostpair, HostIfacePair{hostnames[i], string(interfacename)})
						// create_the_connections
					} else {
						panic("Mismatch with the destination.")
					}
				}
			}
			fmt.Println(interfacename)
		}
		fmt.Println(hostnames[i])
		fmt.Println(conf.Hosts[hostnames[i]].Interfaces)
		fmt.Println(createResults[i].Config.Name)
		fmt.Println(dockerResults[i].Netnsid)
	}
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf(
		"Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}

func main() {
	fmt.Println("==== Parsing YAML file ====")
	config, _ := parser.ParseYAMLFile("simple.yaml")
	parser.ConnectNetworkconfig(config)

	t := task.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  task.Pending,
		Image:  "Image-1",
		Memory: 1024,
		Disk:   1,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Name:  "worker-1",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name:   "Node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "worker",
	}

	fmt.Printf("node: %v\n", n)

	fmt.Printf("create a test container\n")
	/*
		dockerTask, createResult := createContainer()
		if createResult.Error != nil {
			fmt.Printf("%v", createResult.Error)
			os.Exit(1)
		}

		time.Sleep(time.Second * 5)

		fmt.Printf("stopping container %s\n", createResult.ContainerId)
		_ = stopContainer(dockerTask, createResult.ContainerId)
	*/

	dockerTasks, createResults, hostnames := createContainersFromConfig(config)
	// connections := createConnectionsFromNetwork(config)
	createConnectionsFromNetworkConfig(config, dockerTasks, createResults, hostnames)

	var i string

	fmt.Scanln(&i)

	for i := range dockerTasks {
		createResult := createResults[i]
		dockerTask := dockerTasks[i]

		if createResult.Error != nil {
			fmt.Printf("%v", createResult.Error)
			os.Exit(1)
		}

		fmt.Printf("stopping container %s\n", createResult.ContainerId)
		_ = stopContainer(dockerTask, createResult.ContainerId)
		exec.Command("sudo", "ip", "netns", "del", createResult.Netnsid)
	}
}
