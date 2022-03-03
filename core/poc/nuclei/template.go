package nuclei

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
)

func NucleiToMsg(t *templates.Template) string {
	id := t.ID
	message := fmt.Sprintf("Loading nuclei PoC %s[%s] (%s)",
		aurora.Bold(t.Info.Name).String(),
		id,
		aurora.BrightYellow("@"+t.Info.Authors.String()).String())
	return message
}
