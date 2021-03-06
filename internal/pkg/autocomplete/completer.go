package autocomplete

import (
	"fmt"
	"strings"

	"github.com/SierraSoftworks/git-tool/internal/pkg/di"
	"github.com/SierraSoftworks/git-tool/pkg/models"
)

type Completer struct {
	Filter string

	repos []models.Repo
}

func NewCompleter(filter string) *Completer {
	return &Completer{
		Filter: filter,
	}
}

func (c *Completer) Fixed(options ...string) {
	for _, opt := range options {
		c.complete(opt)
	}
}

func (c *Completer) getRepos() []models.Repo {
	if c.repos != nil {
		return c.repos
	}

	rs, err := di.GetMapper().GetRepos()
	if err != nil {
		return []models.Repo{}
	}

	c.repos = rs
	return rs
}

func (c *Completer) complete(value string) {
	if c.matchesFilter(value) {
		if strings.ContainsAny(value, " \t\n\r") {
			fmt.Fprintf(di.GetOutput(), "'%s'\n", value)
		} else {
			fmt.Fprintln(di.GetOutput(), value)
		}
	}
}

func (c *Completer) matchesFilter(value string) bool {
	return Matches(strings.ToLower(value), strings.ToLower(c.Filter))
}
