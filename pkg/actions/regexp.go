package actions

import (
	"context"
	"regexp"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type RegexpAction struct {
	Src   string `yaml:"src"`
	Exp   string `yaml:"exp"`
	Dst   string `yaml:"dst"`
	Split bool   `yaml:"split"`
}

func (a *RegexpAction) parse(vars *model.Vars) (*RegexpAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	return &RegexpAction{
		Src:   common.ParseTplWithPanic(a.Src, vars),
		Exp:   common.ParseTplWithPanic(a.Exp, vars),
		Dst:   common.ParseTplWithPanic(a.Dst, vars),
		Split: a.Split,
	}, gerr
}

func (a *RegexpAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	newa, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	vRegexp := regexp.MustCompile(newa.Exp)
	if newa.Split {
		vParams := vRegexp.Split(newa.Src, -1)
		common.SetVar(newa.Dst, vParams, vars)
		return "", nil
	}
	vParams := vRegexp.FindStringSubmatch(newa.Src)
	common.SetVar(newa.Dst, vParams[1:], vars)
	return "", nil
}
