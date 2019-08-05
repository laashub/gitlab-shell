package commandargs

import (
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/executable"
)

type CommandType string

type CommandArgs interface {
	Parse() error
	GetArguments() []string
}

func Parse(e *executable.Executable, arguments []string) (CommandArgs, error) {
	var args CommandArgs = &GenericArgs{Arguments: arguments}

	switch e.Name {
	case executable.GitlabShell:
		args = &Shell{Arguments: arguments}
	}

	if err := args.Parse(); err != nil {
		return nil, err
	}

	return args, nil
}
