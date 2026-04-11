package commands

import "sort"

// Command represents a registered view command (e.g., ":index", ":node").
type Command struct {
	Name        string // e.g., "index"
	Aliases     []string
	Description string
}

// Router manages command registration and dispatch.
type Router struct {
	commands []Command
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Register(cmd Command) {
	r.commands = append(r.commands, cmd)
}

// Match returns the command that matches the given input, or nil.
func (r *Router) Match(input string) *Command {
	for i := range r.commands {
		if r.commands[i].Name == input {
			return &r.commands[i]
		}
		for _, alias := range r.commands[i].Aliases {
			if alias == input {
				return &r.commands[i]
			}
		}
	}
	return nil
}

// Complete returns commands that match the given prefix, sorted by name.
func (r *Router) Complete(prefix string) []Command {
	if prefix == "" {
		return r.commands
	}
	var matches []Command
	for _, cmd := range r.commands {
		if len(cmd.Name) >= len(prefix) && cmd.Name[:len(prefix)] == prefix {
			matches = append(matches, cmd)
			continue
		}
		for _, alias := range cmd.Aliases {
			if len(alias) >= len(prefix) && alias[:len(prefix)] == prefix {
				matches = append(matches, cmd)
				break
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})
	return matches
}

// Names returns all registered command names.
func (r *Router) Names() []string {
	names := make([]string, len(r.commands))
	for i, cmd := range r.commands {
		names[i] = cmd.Name
	}
	return names
}
