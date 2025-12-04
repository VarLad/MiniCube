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
	"log"
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

func createContainersFromConfig(conf *parser.NetworkConfig) ([]*task.Docker, map[string]*task.DockerResult) {
	ds := []*task.Docker{}
	results := make(map[string]*task.DockerResult)

	for hostname, host := range conf.Hosts {
		if host.Type == "docker" || host.Type == "frr-docker" {

			image := "alpine:latest"

			cmds := []string{"sleep", "infinity"}

			env := []string{"HELLO=cube"}

			switch host.Type {
			case "ovs-docker":
				image = "debian-ovs"
				cmds = []string{"sh", "-c", "ovs-ctl start; ovs-vsctl add-br br0; sleep infinity"}
			case "frr-docker":
				image = "frrouting/frr:latest"
				env = []string{"FRR_DAEMONS=\"zebra bgpd ospfd\""}
			}

			c := task.Config{
				Name:  string(hostname),
				Image: image,
				Env:   env,
			}

			dc, _ := client.NewClientWithOpts(client.FromEnv)
			d := task.Docker{
				Client: dc,
				Config: c,
			}

			result := d.Run(cmds)
			if result.Error != nil {
				fmt.Printf("%v\n", result.Error)
				return nil, nil
			}
			fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
			ds = append(ds, &d)
			results[hostname] = &result
		}
	}

	return ds, results
}

type HostIfacePair struct {
	Host  string
	Iface string
}

/*
func container_exec(containertype string, name string, super string, extra []string) {

}
*/

func create_connection(hosttype1 string, hosttype2 string, hostname1 string, hostname2 string, ifacename1 string, ifacename2 string, netnsid1 string, netnsid2 string, iplist1 []string, iplist2 []string, routelis1 map[string]parser.Route, routelist2 map[string]parser.Route) {
	fmt.Println()
	fmt.Println("Connection Creation")
	fmt.Println(hostname1, hostname2, ifacename1, ifacename2, netnsid1, netnsid2, iplist1, iplist2)
	fmt.Println()
	veth1 := "veth1"
	veth2 := "veth2"

	// exec.Command("sudo", "docker", "exec", hostname1, "ovs-ctl", "start")
	c := exec.Command("sudo", "ip", "link", "add", veth1, "type", "veth", "peer", "name", veth2)
	_, err := c.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	c = exec.Command("sudo", "ip", "link", "set", veth1, "netns", netnsid1)
	_, err = c.Output()
	if err != nil {
		log.Fatal(err)
	}
	/*
		if hosttype1 == "ovs-docker" {
			c = exec.Command("sudo", "docker", "exec", hostname1, "ovs-vsctl", "add-port", "br0", ifacename1)
			_, err = c.Output()
			if err != nil {
				log.Fatal(err)
			}
		}
	*/
	exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "link", "set", "lo", "up").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "link", "set", veth1, "up").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "link", "set", veth1, "down").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "link", "set", veth1, "name", ifacename1).Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "link", "set", ifacename1, "up").Run()
	for _, ipaddr := range iplist1 {
		exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "addr", "add", ipaddr, "dev", ifacename1).Run()
	}

	exec.Command("sudo", "ip", "link", "set", veth2, "netns", netnsid2).Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "link", "set", "lo", "up").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "link", "set", veth2, "up").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "link", "set", veth2, "down").Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "link", "set", veth2, "name", ifacename2).Run()
	exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "link", "set", ifacename2, "up").Run()
	for _, ipaddr := range iplist2 {
		exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "addr", "add", ipaddr, "dev", ifacename2).Run()
	}
	/*
			if hosttype2 == "ovs-docker" {
				exec.Command("sudo", "docker", "exec", hostname2, "ovs-vsctl", "add-port", "br0", ifacename2).Run()
			}
		}
	*/
	if hosttype1 == "docker" {
		for _, i := range routelis1 {
			exec.Command("sudo", "ip", "netns", "exec", netnsid1, "ip", "route", "add", i.To, "via", i.Via)
		}
	}
	if hosttype2 == "docker" {
		for _, i := range routelist2 {
			exec.Command("sudo", "ip", "netns", "exec", netnsid2, "ip", "route", "add", i.To, "via", i.Via)
		}
	}
}

/*
func create_connection(hostname1 string, hostname2 string, ifacename1 string if) {

}
*/
func createHostConnectionsFromNetworkConfig(conf *parser.NetworkConfig, dockerTasks []*task.Docker, createResults map[string]*task.DockerResult) {
	// fmt.Println("Meow")
	// fmt.Println(conf.Hosts["h3"].Interfaces)
	fmt.Println(conf.Hosts["frr1"].Routes)
	fmt.Println("Meow")
	fmt.Println("----")
	usedhostpair := []HostIfacePair{}
	useddestpair := []HostIfacePair{}
	for _, dockerTask := range dockerTasks {
		hostname := dockerTask.Config.Name
		for interfacename, iface := range conf.Hosts[hostname].Interfaces {
			reversepair := conf.Hosts[iface.DstNode].Interfaces[iface.DstIface]
			if slices.Contains(usedhostpair, HostIfacePair{iface.DstNode, iface.DstIface}) && slices.Contains(useddestpair, HostIfacePair{hostname, interfacename}) {
				if reversepair.DstNode == string(hostname) && reversepair.DstIface == interfacename {
					useddestpair = append(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface})
					usedhostpair = append(usedhostpair, HostIfacePair{hostname, interfacename})
					continue
				} else {
					panic("Mismatch with destination")
				}
			}
			if slices.Contains(usedhostpair, HostIfacePair{hostname, string(interfacename)}) {
				panic("The same interface name or host name has been likely declared more than once.")
			} else {
				if slices.Contains(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface}) {
					panic("This interface is already in use.")
				} else {
					if reversepair.DstNode == hostname && reversepair.DstIface == interfacename {
						useddestpair = append(useddestpair, HostIfacePair{iface.DstNode, iface.DstIface})
						usedhostpair = append(usedhostpair, HostIfacePair{hostname, interfacename})
						create_connection(conf.Hosts[hostname].Type, conf.Hosts[iface.DstNode].Type, hostname, iface.DstNode, interfacename, iface.DstIface, createResults[hostname].Netnsid, createResults[iface.DstNode].Netnsid, iface.Addresses, reversepair.Addresses, conf.Hosts[hostname].Routes, conf.Hosts[iface.DstNode].Routes)
					} else {
						fmt.Println(reversepair.DstNode)
						fmt.Println(hostname)
						fmt.Println(reversepair.DstIface)
						fmt.Println(interfacename)
						panic("Mismatch with the destination.")
					}
				}
			}
			fmt.Println(interfacename)
		}
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
	config, _ := parser.ParseYAMLFile("simple_frr.yaml")
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

	dockerTasks, createResults := createContainersFromConfig(config)
	// connections := createConnectionsFromNetwork(config)
	createHostConnectionsFromNetworkConfig(config, dockerTasks, createResults)

	var i string

	fmt.Scanln(&i)

	for _, dockerTask := range dockerTasks {
		createResult := createResults[dockerTask.Config.Name]
		if createResult.Error != nil {
			fmt.Printf("%v", createResult.Error)
			os.Exit(1)
		}

		fmt.Printf("stopping container %s\n", createResult.ContainerId)
		_ = stopContainer(dockerTask, createResult.ContainerId)
		exec.Command("sudo", "ip", "netns", "del", createResult.Netnsid)
	}
}
