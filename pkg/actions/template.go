package actions

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type TemplateAction struct {
	Src   string `yaml:"src"`
	Dest  string `yaml:"dest"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
	Mode  string `yaml:"mode"`
}

func (a *TemplateAction) parse(vars *model.Vars) (*TemplateAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &TemplateAction{
		Src:   common.ParseTplWithPanic(a.Src, vars),
		Dest:  common.ParseTplWithPanic(a.Dest, vars),
		Owner: common.ParseTplWithPanic(a.Owner, vars),
		Group: common.ParseTplWithPanic(a.Group, vars),
		Mode:  common.ParseTplWithPanic(a.Mode, vars),
	}, gerr
}

func (a *TemplateAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	tpl, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, parseAction.Src))
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %s", err)
	}
	data, err := common.ParseTpl(string(tpl), vars)
	if err != nil {
		return "", fmt.Errorf("failed to parse source file: %s", err)
	}
	buf := bytes.NewBufferString(data)
	mode := parseAction.Mode
	if mode == "" {
		mode = "0644"
	}
	err = conn.CopyFile(ctx, buf, int64(len(data)), parseAction.Dest, mode)
	if err != nil {
		return "", fmt.Errorf("failed to copy file %q: %s", parseAction.Src, err)
	}

	if parseAction.Owner != "" && parseAction.Group != "" {
		output, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(
				fmt.Sprintf("chown %s:%s %s", parseAction.Owner, parseAction.Group, parseAction.Dest),
			), nil
		})
		if err != nil {
			return output, fmt.Errorf(
				"failed to set the file owner on %q to %s:%s: %s",
				parseAction.Dest, parseAction.Owner, parseAction.Group, err,
			)
		}
	}

	return "", nil
}
