// Package commands
package commands

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/client/promt/responses"
)

type (
	CommandRespond interface {
		GetToken() (string, bool)
	}

	Command struct {
		name     string
		help     string
		execFunc func(context.Context, ...string) *responses.Respond
	}
)

func New(name string, help string, f func(context.Context, ...string) *responses.Respond) *Command {
	return &Command{
		name:     name,
		help:     help,
		execFunc: f,
	}
}

func (c *Command) String() string {
	return c.name
}

func (c *Command) PrintHelp() string {
	return c.help
}
func (c *Command) Exec(ctx context.Context, s ...string) *responses.Respond {
	if c.execFunc != nil {
		return c.execFunc(ctx, s...)
	}
	return nil
}
