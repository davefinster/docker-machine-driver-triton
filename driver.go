package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
    "github.com/joyent/gosign/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

func main() {
	plugin.RegisterDriver(&Driver{})
}

type Driver struct {
	*drivers.BaseDriver
	TritonUsername	string
	TritonKeyPath	string
	TritonKeyId		string
	TritonEndpoint	string

	Package 		string
	Image			string
	Network			string

	SSHUsername 	string

	MachineId		string
}

const (
	defaultPackage = "k4-general-kvm-3.75G"
	defaultImage = "698a8146-d6d9-4352-99fe-6557ebce5661"
	defaultNetwork = "f7ed95d3-faaf-43ef-9346-15644403b963"
	defaultTritonEndpoint = "https://us-sw-1.api.joyent.com"
	defaultImageSshUsername = "ubuntu"
)

// GetCreateFlags registers the flags this d adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "TRITON_USERNAME",
			Name:   "triton-username",
			Usage:  "REQUIRED: Triton username",
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_KEY_PATH",
			Name:   "triton-key-path",
			Usage:  "REQUIRED: Path to PEM private key",
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_KEY_ID",
			Name:   "triton-key-id",
			Usage:  "REQUIRED: ID of PEM private key",
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_ENDPOINT",
			Name:   "triton-endpoint",
			Usage:  "triton cloudapi HTTP endpoint (default https://us-sw-1.api.joyent.com)",
			Value:  defaultTritonEndpoint,
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_PACKAGE",
			Name:   "triton-package",
			Usage:  "triton machine package (default:k4-general-kvm-3.75G)",
			Value: 	defaultPackage,
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_IMAGE",
			Name:   "triton-image",
			Usage:  "machine image to use (default:698a8146-d6d9-4352-99fe-6557ebce5661)",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_NETWORK",
			Name:   "triton-network",
			Usage:  "triton network name (default:f7ed95d3-faaf-43ef-9346-15644403b963)",
			Value:  defaultNetwork,
		},
		mcnflag.StringFlag{
			EnvVar: "TRITON_IMAGE_SSH_USERNAME",
			Name:   "triton-image-ssh-username",
			Usage:  "triton network name (default:ubuntu)",
			Value:  defaultImageSshUsername,
		},
	}
}

var tritonClient *cloudapi.Client

func (d *Driver) client() *cloudapi.Client {
	if tritonClient == nil {
		keyData, _ := ioutil.ReadFile(d.TritonKeyPath)
		userAuth, err := auth.NewAuth(d.TritonUsername, string(keyData), "rsa-sha256")
		creds := &auth.Credentials{
	        UserAuthentication: userAuth,
	        SdcKeyId:           d.TritonKeyId,
	        SdcEndpoint:        auth.Endpoint{URL: d.TritonEndpoint},
	    }
	    tritonClient = cloudapi.New(client.NewClient(
	        creds.SdcEndpoint.URL,
	        cloudapi.DefaultAPIVersion,
	        creds,
	        nil,
	    ))
	    _, err = tritonClient.ListKeys()
	    if err != nil {
	    	log.Errorf("Error authenticating %v", err)
		}
	}
	return tritonClient
}


// NewDriver instantiates a new driver with hostName into storePath
func NewDriver(hostName, storePath string) drivers.Driver {
	d := &Driver{
		TritonEndpoint:			defaultTritonEndpoint,
		Package:				defaultPackage,
		Image: 					defaultImage,
		Network: 				defaultNetwork,
		SSHUsername: 			defaultImageSshUsername,
		BaseDriver: &drivers.BaseDriver{
			SSHPort:     22,
			SSHUser:     defaultImageSshUsername,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
	return d
}

// SetConfigFromFlags implements interface method for parsing cmdline args
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.TritonUsername = flags.String("triton-username")
	d.TritonKeyPath = flags.String("triton-key-path")
	d.TritonKeyId = flags.String("triton-key-id")
	d.TritonEndpoint = flags.String("triton-endpoint")

	d.Package = flags.String("triton-package")
	d.Image = flags.String("triton-image")
	d.Network = flags.String("triton-network")

	d.SSHUsername = flags.String("triton-image-ssh-username")

	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) GetIP() (string, error) {
	machine, err := d.client().GetMachine(d.MachineId)
	if (err != nil) {
		return "", err
	}
	return machine.IPs[0], nil
}

// GetSSHHostname aliases GetIP
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUsername
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%v:2376", ip), nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "triton"
}

func (d *Driver) createTritonInstance() *cloudapi.Machine {
	machine, err := d.client().CreateMachine(cloudapi.CreateMachineOpts{
		Name: d.MachineName,
		Package: d.Package,
		Image: d.Image,
		Networks: []string{d.Network},
		FirewallEnabled: false,
	})
	if (err != nil) {
		log.Debugf("Error creating Triton instance: %v", err)
		//the request has failed - wait a minute polling getmachines to see if it shows up
		hasMachine := false
		checkCount := 0
		machineFilter := cloudapi.NewFilter()
		machineFilter.Set("name", d.MachineName)
		for ((hasMachine == false) && (checkCount < 12)) {
			machines, _ := d.client().ListMachines(machineFilter)
			for _, m := range machines {
				log.Debugf("Checking machine with name", m.Name, "compared with", d.MachineName)
				if (m.Name == d.MachineName) {
					machine = &m
					hasMachine = true
					log.Debugf("Found machine that was meant to be created")
				}
			}
			checkCount += 1
			time.Sleep(5 * time.Second)
		}
	}
	if (machine == nil) {
		log.Debugf("Launching another instance creation request")
		return d.createTritonInstance()
	}
	return machine
} 

func (d *Driver) Create() error {
	machine := d.createTritonInstance()
	machineRunning := false
	for (machineRunning == false) {
		machine, _ = d.client().GetMachine(machine.Id)
		if (machine.State == "running") {
			machineRunning = true
		}
		time.Sleep(5 * time.Second)
	}
	d.MachineId = machine.Id
	keyData, err := ioutil.ReadFile(d.TritonKeyPath)
	if (err != nil) {
		return err
	}
    err = ioutil.WriteFile(d.GetSSHKeyPath(), keyData, 0500)
    if (err != nil) {
		return err
	}
	return nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	machine, err := d.client().GetMachine(d.MachineId)
	if err != nil {
		log.Infof("Failed fetching server %v. (is it dead?) error: %v", d.MachineId, err)
		return state.None, nil
	}
	log.Debugf("server.status: %v", machine.State)
	switch machine.State {
	case "provisioning":
		return state.Starting, nil
	case "failed":
		return state.None, nil
	case "running":
		return state.Running, nil
	case "stopping":
		return state.Stopped, nil
	case "stopped":
		return state.Stopped, nil
	}
	return state.None, nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	st, err := d.GetState()
	if st == state.None {
		return nil
	} else if err != nil {
		return fmt.Errorf("Failed fetching state: %v", err)
	}
	return d.client().DeleteMachine(d.MachineId)
}

// Start a host
func (d *Driver) Start() error {
	err := d.client().StartMachine(d.MachineId)
	if err != nil {
		return fmt.Errorf("Failed starting server: %v - %v", d.MachineId, err)
	}
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	err := d.client().StopMachine(d.MachineId)
	if err != nil {
		return fmt.Errorf("Failed stopping server: %v - %v", d.MachineId, err)
	}
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return d.client().RebootMachine(d.MachineId)
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return d.Stop()
}

