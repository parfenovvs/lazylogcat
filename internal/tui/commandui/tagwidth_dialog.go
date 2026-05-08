package commandui

import (
	"fmt"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

type TagWidthSelectMsg struct {
	Width int
}

func newTagWidthSelect(prefs model.OutputPrefs) NumberSelectorModel {
	return newNumberSelect(NumberSelectorConfig{
		Title:  "Tag width",
		Footer: "Set the width of the Tag column.\nEmpty or 0 for auto width.",
		Value:  prefs.TagWidth,
		Max:    util.MaxTagWidth,
		ValidateFn: func(value int) error {
			if value < 0 {
				return fmt.Errorf("tag width cannot be negative")
			}
			if value > util.MaxTagWidth {
				return fmt.Errorf("tag width cannot exceed %d", util.MaxTagWidth)
			}
			return nil
		},
	})
}
