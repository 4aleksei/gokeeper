// Package command
package command

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/client/promt/responses"
	"github.com/4aleksei/gokeeper/internal/client/service"
)

type (
	CommandRespond interface {
		GetToken() (string, bool)
	}

	Command struct {
		srv      *service.HandleService
		name     string
		help     string
		execFunc func(context.Context, *service.HandleService, ...string) *responses.Respond
	}
)

func New(srv *service.HandleService, name string, help string, f func(context.Context, *service.HandleService, ...string) *responses.Respond) *Command {
	return &Command{
		name:     name,
		help:     help,
		execFunc: f,
		srv:      srv,
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
		return c.execFunc(ctx, c.srv, s...)
	}
	return nil
}
