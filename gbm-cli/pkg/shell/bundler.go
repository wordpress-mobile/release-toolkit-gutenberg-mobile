package shell

type BundlerCmds interface {
	Install(...string) error
	PodInstall(...string) error
}

func (c *client) PodInstall(args ...string) error {
	podInstall := append([]string{"exec", "pod", "install"}, args...)
	return c.cmd(podInstall...)
}
