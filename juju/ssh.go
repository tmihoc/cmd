package main

import (
	"errors"
	"fmt"
	"launchpad.net/gnuflag"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/log"
	"launchpad.net/juju-core/state"
	"os/exec"
)

// SSHCommand is responsible for launching a ssh shell on a given unit or machine.
type SSHCommand struct {
	SSHCommon
}

// SSHCommon provides common methods for SSHCommand and SCPCommand.
type SSHCommon struct {
	EnvName string
	Target  string
	Args    []string
	*juju.Conn
}

func (c *SSHCommand) Info() *cmd.Info {
	return &cmd.Info{"ssh", "", "launch an ssh shell on a given unit or machine", ""}
}

func (c *SSHCommand) SetFlags(f *gnuflag.FlagSet) {
	addEnvironFlags(&c.EnvName, f)
}

func (c *SSHCommand) Init(args []string) error {
	if len(args) == 0 {
		return errors.New("no service name specified")
	}
	c.Target, c.Args = args[0], args[1:]
	return nil
}

// Run resolves c.Target to a machine, to the address of a i
// machine or unit forks ssh passing any arguments provided.
func (c *SSHCommand) Run(ctx *cmd.Context) error {
	var err error
	c.Conn, err = juju.NewConnFromName(c.EnvName)
	if err != nil {
		return err
	}
	defer c.Close()
	host, err := c.hostFromTarget(c.Target)
	if err != nil {
		return err
	}
	args := []string{"-l", "ubuntu", "-t", "-o", "StrictHostKeyChecking no", "-o", "PasswordAuthentication no", host}
	args = append(args, c.Args...)
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = ctx.Stdin
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	c.Close()
	return cmd.Run()
}

func (c *SSHCommon) hostFromTarget(target string) (string, error) {
	// is the target the id of a machine ?
	if state.IsMachineId(target) {
		log.Printf("cmd/juju: looking up address for machine %s...", target)
		// TODO(dfc) maybe we should have machine.PublicAddress() ?
		return c.machinePublicAddress(target)
	}
	// maybe the target is a unit ?
	if state.IsUnitName(target) {
		log.Printf("cmd/juju: Looking up address for unit %q...", c.Target)
		unit, err := c.State.Unit(target)
		if err != nil {
			return "", err
		}
		addr, ok := unit.PublicAddress()
		if !ok {
			return "", fmt.Errorf("unit %q has no public address", unit)
		}
		return addr, nil
	}
	return "", fmt.Errorf("unknown unit or machine %q", target)
}

func (c *SSHCommon) machinePublicAddress(id string) (string, error) {
	machine, err := c.State.Machine(id)
	if err != nil {
		return "", err
	}
	// wait for instance id
	w := machine.Watch()
	for _ = range w.Changes() {
		if instid, ok := machine.InstanceId(); ok {
			w.Stop()
			inst, err := c.Environ.Instances([]state.InstanceId{instid})
			if err != nil {
				return "", err
			}
			return inst[0].WaitDNSName()
		}
		// BUG(dfc) this does not refresh the machine, so
		// this loop will loop forever if it gets to this point.
		// https://bugs.launchpad.net/juju-core/+bug/1130051
	}
	// oops, watcher closed before we could get an answer
	return "", w.Stop()
}
