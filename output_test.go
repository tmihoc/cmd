// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENSE file for details.

package cmd_test

import (
	"github.com/juju/gnuflag"
	gc "gopkg.in/check.v1"

	"github.com/juju/cmd/v3"
	"github.com/juju/cmd/v3/cmdtesting"
)

// OutputCommand is a command that uses the output.go formatters.
type OutputCommand struct {
	cmd.CommandBase
	out   cmd.Output
	value interface{}
}

func (c *OutputCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "output",
		Args:    "<something>",
		Purpose: "I like to output",
		Doc:     "output",
	}
}

func (c *OutputCommand) SetFlags(f *gnuflag.FlagSet) {
	formatters := make(map[string]cmd.Formatter, len(cmd.DefaultFormatters))
	for k, v := range cmd.DefaultFormatters {
		formatters[k] = v.Formatter
	}
	c.out.AddFlags(f, "smart", formatters)
}

func (c *OutputCommand) Init(args []string) error {
	return cmd.CheckEmpty(args)
}

func (c *OutputCommand) Run(ctx *cmd.Context) error {
	if value, ok := c.value.(overrideFormatter); ok {
		return c.out.WriteFormatter(ctx, value.formatter, value.value)
	}
	return c.out.Write(ctx, c.value)
}

type overrideFormatter struct {
	formatter cmd.Formatter
	value     interface{}
}

// use a struct to control field ordering.
var defaultValue = struct {
	Juju   int
	Puppet bool
}{1, false}

var outputTests = map[string][]struct {
	value  interface{}
	output string
}{
	"": {
		{nil, ""},
		{"", ""},
		{1, "1\n"},
		{-1, "-1\n"},
		{1.1, "1.1\n"},
		{10000000, "10000000\n"},
		{true, "True\n"},
		{false, "False\n"},
		{"hello", "hello\n"},
		{"\n\n\n", "\n\n\n\n"},
		{"foo: bar", "foo: bar\n"},
		{[]string{}, ""},
		{[]string{"blam", "dink"}, "blam\ndink\n"},
		{map[interface{}]interface{}{"foo": "bar"}, "foo: bar\n"},
		{overrideFormatter{cmd.FormatSmart, "abc\ndef"}, "abc\ndef\n"},
	},
	"smart": {
		{nil, ""},
		{"", ""},
		{1, "1\n"},
		{-1, "-1\n"},
		{1.1, "1.1\n"},
		{10000000, "10000000\n"},
		{true, "True\n"},
		{false, "False\n"},
		{"hello", "hello\n"},
		{"\n\n\n", "\n\n\n\n"},
		{"foo: bar", "foo: bar\n"},
		{[]string{}, ""},
		{[]string{"blam", "dink"}, "blam\ndink\n"},
		{map[interface{}]interface{}{"foo": "bar"}, "foo: bar\n"},
		{overrideFormatter{cmd.FormatSmart, "abc\ndef"}, "abc\ndef\n"},
	},
	"json": {
		{nil, "null\n"},
		{"", `""` + "\n"},
		{1, "1\n"},
		{-1, "-1\n"},
		{1.1, "1.1\n"},
		{10000000, "10000000\n"},
		{true, "true\n"},
		{false, "false\n"},
		{"hello", `"hello"` + "\n"},
		{"\n\n\n", `"\n\n\n"` + "\n"},
		{"foo: bar", `"foo: bar"` + "\n"},
		{[]string{}, `[]` + "\n"},
		{[]string{"blam", "dink"}, `["blam","dink"]` + "\n"},
		{defaultValue, `{"Juju":1,"Puppet":false}` + "\n"},
		{overrideFormatter{cmd.FormatSmart, "abc\ndef"}, "abc\ndef\n"},
		{overrideFormatter{cmd.FormatJson, struct{}{}}, "{}\n"},
	},
	"yaml": {
		{nil, ""},
		{"", `""` + "\n"},
		{1, "1\n"},
		{-1, "-1\n"},
		{1.1, "1.1\n"},
		{10000000, "10000000\n"},
		{true, "true\n"},
		{false, "false\n"},
		{"hello", "hello\n"},
		{"\n\n\n", "|2+\n"},
		{"foo: bar", "'foo: bar'\n"},
		{[]string{}, "[]\n"},
		{[]string{"blam", "dink"}, "- blam\n- dink\n"},
		{defaultValue, "juju: 1\npuppet: false\n"},
		{overrideFormatter{cmd.FormatSmart, "abc\ndef"}, "abc\ndef\n"},
		{overrideFormatter{cmd.FormatYaml, struct{}{}}, "{}\n"},
	},
}

func (s *CmdSuite) TestOutputFormat(c *gc.C) {
	for format, tests := range outputTests {
		c.Logf("format %s", format)
		var args []string
		if format != "" {
			args = []string{"--format", format}
		}
		for i, t := range tests {
			c.Logf("  test %d", i)
			ctx := cmdtesting.Context(c)
			result := cmd.Main(&OutputCommand{value: t.value}, ctx, args)
			c.Check(result, gc.Equals, 0)
			c.Check(bufferString(ctx.Stdout), gc.Equals, t.output)
			c.Check(bufferString(ctx.Stderr), gc.Equals, "")
		}
	}
}

func (s *CmdSuite) TestUnknownOutputFormat(c *gc.C) {
	ctx := cmdtesting.Context(c)
	result := cmd.Main(&OutputCommand{}, ctx, []string{"--format", "cuneiform"})
	c.Check(result, gc.Equals, 2)
	c.Check(bufferString(ctx.Stdout), gc.Equals, "")
	c.Check(bufferString(ctx.Stderr), gc.Matches, ".*: unknown format \"cuneiform\"\n")
}

// Py juju allowed both --format json and --format=json. This test verifies that juju is
// being built against a version of the gnuflag library (rev 14 or above) that supports
// this argument format.
// LP #1059921
func (s *CmdSuite) TestFormatAlternativeSyntax(c *gc.C) {
	ctx := cmdtesting.Context(c)
	result := cmd.Main(&OutputCommand{}, ctx, []string{"--format=json"})
	c.Assert(result, gc.Equals, 0)
	c.Assert(bufferString(ctx.Stdout), gc.Equals, "null\n")
}

func (s *CmdSuite) TestFormatters(c *gc.C) {
	typeFormatters := cmd.DefaultFormatters
	formatters := typeFormatters.Formatters()

	c.Assert(len(typeFormatters), gc.Equals, len(formatters))
	for k := range typeFormatters {
		_, ok := formatters[k]
		c.Assert(ok, gc.Equals, true)
	}
}
