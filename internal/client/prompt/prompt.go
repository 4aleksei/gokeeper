// Package prompt
package prompt

import (
	"context"
	"strings"

	"github.com/4aleksei/gokeeper/internal/client/prompt/command"
	"github.com/4aleksei/gokeeper/internal/client/prompt/responses"
	"github.com/4aleksei/gokeeper/internal/client/service"
	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/pterm/pterm"
)

type (
	PromptCommand interface {
		String() string
		PrintHelp() string
		Exec(context.Context, ...string) *responses.Respond
	}

	Prompt struct {
		commands map[string]PromptCommand
		token    string
		name     string
	}
)

func AddCommand(comm *command.Command) func(*Prompt) {

	return func(p *Prompt) {

		p.commands[comm.String()] = comm

	}
}

func New(options ...func(*Prompt)) *Prompt {
	prm := &Prompt{
		commands: make(map[string]PromptCommand),
	}
	options = append(options, AddCommand(command.New(nil, "Help", "list all commands", func(context.Context, *service.HandleService, ...string) *responses.Respond {
		pterm.Println()
		for _, o := range prm.commands {
			pterm.Info.Printfln("Command:  %s , Usage: %s", o.String(), o.PrintHelp())
			pterm.Println()
		}
		return nil
	})))

	for _, o := range options {
		o(prm)
	}
	return prm
}

func (p *Prompt) Loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			var result string
			if p.name != "" {
				result, _ = pterm.DefaultInteractiveTextInput.Show(p.name + ">")
			} else {
				result, _ = pterm.DefaultInteractiveTextInput.Show(">")
			}

			fields := strings.Fields(result)
			if len(fields) < 1 {
				continue
			}
			v, ok := p.commands[fields[0]]
			if !ok {
				pterm.Printfln("Unknow command %s\n", fields[0])
				continue
			}

			pterm.Printfln("Execute command %s", v)

			var arg []string
			arg = append(arg, p.token)
			arg = append(arg, fields[1:]...)
			p.getResponse(v.Exec(ctx, arg...))
		}
	}
}

func (p *Prompt) getResponse(resp *responses.Respond) {
	if resp == nil {
		return
	}

	err := resp.GetError()
	if err != nil {
		pterm.Printfln("Command with %v", err)
		return
	}

	if t, ok := resp.GetToken(); ok {
		p.token = t
		p.name = resp.GetMetaData()
		return
	}

	if data, ok := resp.GetData(); ok {
		pterm.Printfln("Load Data :%s", data)
		pterm.Printfln("Metadata :%s", resp.GetMetaData())
		pterm.Printfln("Typedata :%s", store.GetStringType(resp.GetType()))
		return
	}

	if data, ok := resp.GetUUID(); ok {
		pterm.Printfln("Data UUID :%s", data)
		return
	}
}
