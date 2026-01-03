package g2cmd

import (
	"fmt"

	"github.com/atcharles/gof/v2/g2util"
)

// BuildInfoInstance ...
var BuildInfoInstance = make(g2util.Map)

// BuildInfoInstanceAdd ...
func BuildInfoInstanceAdd(info g2util.Map) {
	for k, v := range info {
		BuildInfoInstance[k] = v
	}
}

func showVersion() {
	verStr := `------------------------------版本信息------------------------------
Application: {{.appName}}
Version: {{.version}}
BuildTime: {{.buildTime}}
{{.goVersion}}
git: {{.gitBranch}} - {{.gitHash}}`
	fmt.Println(g2util.TextTemplateMustParse(verStr, BuildInfoInstance))
}
