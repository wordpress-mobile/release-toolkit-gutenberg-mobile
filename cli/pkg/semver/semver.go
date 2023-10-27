package semver

import "fmt"

type semver struct {
	Major int
	Minor int
	Patch int
}

type SemVer interface {
	String() string
	Vstring() string
	PriorVersion() string
	IsScheduledRelease() bool
	IsPatchRelease() bool
	Parse(version string) error
}

func NewSemVer(version string) (SemVer, error) {
	s := &semver{}
	err := s.Parse(version)
	return s, err
}

func (s *semver) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

func (s *semver) Vstring() string {
	return fmt.Sprintf("v%d.%d.%d", s.Major, s.Minor, s.Patch)
}

func (s *semver) PriorVersion() string {
	if s.IsPatchRelease() {
		return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch-1)
	}
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor-1, 0)
}

func (s *semver) IsScheduledRelease() bool {
	return s.Patch == 0
}

func (s *semver) IsPatchRelease() bool {
	return s.Patch > 0
}

func (s *semver) Parse(version string) error {
	if version[0] == 'v' {
		version = version[1:]
	}
	_, err := fmt.Sscanf(version, "%d.%d.%d", &s.Major, &s.Minor, &s.Patch)
	return err
}
