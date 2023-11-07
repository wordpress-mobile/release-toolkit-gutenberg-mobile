package shell

type RakeCmds interface {
	Dependencies() error
}

func (c *client) Dependencies() error {
	return c.cmd("dependencies")
}
