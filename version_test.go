// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENSE file for details.

package cmd_test

import (
	"fmt"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/cmd/v3"
	"github.com/juju/cmd/v3/cmdtesting"
)

type VersionSuite struct {
	testing.LoggingSuite
}

var _ = gc.Suite(&VersionSuite{})

type versionDetail struct {
	Version       string `json:"version"`
	GitCommitHash string `json:"git-commit-hash"`
	GitTreeState  string `json:"git-tree-state"`
}

func (s *VersionSuite) TestVersion(c *gc.C) {
	const version = "999.888.777"

	ctx := cmdtesting.Context(c)
	code := cmd.Main(cmd.NewVersionCommand(version, nil), ctx, nil)
	c.Check(code, gc.Equals, 0)
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "")
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, version+"\n")
}

func (s *VersionSuite) TestVersionExtraArgs(c *gc.C) {
	ctx := cmdtesting.Context(c)
	code := cmd.Main(cmd.NewVersionCommand("xxx", nil), ctx, []string{"foo"})
	c.Check(code, gc.Equals, 2)
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, "")
	c.Assert(cmdtesting.Stderr(ctx), gc.Matches, "ERROR unrecognized args.*\n")
}

func (s *VersionSuite) TestVersionJson(c *gc.C) {
	const version = "999.888.777"

	ctx := cmdtesting.Context(c)
	code := cmd.Main(cmd.NewVersionCommand(version, nil), ctx, []string{"--format", "json"})
	c.Check(code, gc.Equals, 0)
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "")
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, fmt.Sprintf("%q\n", version))
}

func (s *VersionSuite) TestVersionDetailJson(c *gc.C) {
	const version = "999.888.777"
	detail := versionDetail{
		Version:       version,
		GitCommitHash: "46f1a0bd5592a2f9244ca321b129902a06b53e03",
		GitTreeState:  "dirty",
	}

	ctx := cmdtesting.Context(c)
	code := cmd.Main(cmd.NewVersionCommand(version, detail), ctx, []string{"--all", "--format", "json"})
	c.Check(code, gc.Equals, 0)
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "")
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, `
{"version":"999.888.777","git-commit-hash":"46f1a0bd5592a2f9244ca321b129902a06b53e03","git-tree-state":"dirty"}
`[1:])
}
