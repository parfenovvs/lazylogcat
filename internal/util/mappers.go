package util

import (
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
)

func FilterFromConfig(c *config.Config) model.Filter {
	return model.Filter{
		PackageName: c.Session.Pkg,
		Tag:         c.Session.Tag,
		Text:        c.Session.Txt,
		Level:       model.LvlV,
	}
}

func FormatFromConfig(c *config.Config) model.Format {
	return model.NewFormat(c.Prefs.Format, c.Prefs.Modifiers...)
}
