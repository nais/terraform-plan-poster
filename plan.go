package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

type Change struct {
	Actions         []string       `json:"actions"`
	Before          map[string]any `json:"before"`
	BeforeSensitive map[string]any `json:"before_sensitive"`
	After           map[string]any `json:"after"`
	AfterSensitive  map[string]any `json:"after_sensitive"`
	AfterUnknown    map[string]any `json:"after_unknown"`
}

type ResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Change        Change `json:"change"`
}

type Plan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

func main() {
	f, err := os.Open("plan-gcp.json")
	if err != nil {
		log.Fatal(err)
	}

	var plan Plan
	if err := json.NewDecoder(f).Decode(&plan); err != nil {
		log.Fatal(err)
	}

outer:
	for _, rc := range plan.ResourceChanges {
		for _, a := range rc.Change.Actions {
			switch a {
			case "update":
				fallthrough
			case "create":
				actions := strings.Join(rc.Change.Actions, ",")
				title := fmt.Sprintf("%s (**%s**)", rc.Address, actions)
				opts := cmp.FilterPath(ignorePath("rule.*.match.*.versioned_expr"), cmp.Ignore())

				diff := code(cmp.Diff(rc.Change.Before, rc.Change.After, opts))

				fmt.Print(wrap(title, diff))
				continue outer
			}
		}
	}
}

func wrap(title, details string) string {
	return fmt.Sprintf(`
<details>
<summary>
%s
</summary>
%s
</details>
	`, title, details)
}

func code(in string) string {
	return fmt.Sprintf("\n```diff\n%s\n```\n", in)
}

func ignorePath(path string) func(p cmp.Path) bool {
	return func(p cmp.Path) bool {
		s := ""
		wide := ""
		for _, pe := range p {
			switch pe := pe.(type) {
			case cmp.MapIndex:
				s += "." + pe.Key().String()
				wide += "." + pe.Key().String()
			case cmp.SliceIndex:
				s += "." + strconv.Itoa(pe.Key())
				wide += ".*"
			}
		}

		// fmt.Println("ASD", s)
		return s == "."+path || wide == "."+path
	}
}
