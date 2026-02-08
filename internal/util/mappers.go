package util

import (
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
)

func FilterFromConfig(c *config.Config) model.Filter {
	return model.Filter{
		PackageName: c.Filter.Pkg.Value,
		Tag:         c.Filter.Tag.Value,
		Text:        c.Filter.Txt.Value,
		Level:       model.LvlV,
	}
}

func ColorFromConfig(c *config.Config) bool {
	if c.Display.Color == nil {
		return true
	}
	return *c.Display.Color
}
