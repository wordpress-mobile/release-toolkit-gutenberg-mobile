package shell

type NpmCmds interface {
	Install(...string) error
	Ci() error
	Run(...string) error
}

func (c *client) Ci() error {
	return c.cmd("ci")
}

func (c *client) Run(args ...string) error {
	run := append([]string{"run"}, args...)
	return c.cmd(run...)
}
