package shell

type NpmCmds interface {
	Install(...string) error
	Ci() error
	Run(...string) error
	RunIn(string, ...string) error
	Version(string) error
	VersionIn(string, string) error
}

func (c *client) Ci() error {
	return c.cmd("ci")
}

func (c *client) Run(args ...string) error {
	run := append([]string{"run"}, args...)
	return c.cmd(run...)
}

func (c *client) RunIn(path string, args ...string) error {
	run := append([]string{"run"}, args...)
	return c.cmdInPath(path, run...)
}

func (c *client) Version(version string) error {
	// Let's not add the tag by default.
	// If we need it we should consider a different function.
	versionCmd := []string{"version", version, "--no-git-tag=false"}
	return c.cmd(versionCmd...)
}

func (c *client) VersionIn(packagePath, version string) error {
	versionCmd := []string{"version", version, "--no-git-tag=false"}
	return c.cmdInPath(packagePath, versionCmd...)
}
