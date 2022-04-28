package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

type ResourceChange struct {
	action  string
	details string
}

func main() {
	segmentPattern := regexp.MustCompile(`^.*# (.*) will be (.*)$`)
	endPattern := regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.`)

	changes := make(map[string]*ResourceChange)
	address := ""
	var end [][]byte
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		end = endPattern.FindSubmatch(line)
		if end != nil {
			if len(end) != 4 {
				log.Fatalf("invalid segment separator: %s", line)
			}
			break
		}
		segment := segmentPattern.FindSubmatch(line)

		// means we've found a separator
		//  # google_compute_security_policy.policy["gw-ekstern-dev-nav-no"] will be updated in-place
		if segment != nil {
			if len(segment) != 3 {
				log.Fatalf("invalid segment separator: %s", line)
			}
			address = string(segment[1])
			changes[address] = &ResourceChange{
				action: string(segment[2]),
			}
		}

		if address != "" {
			changes[address].details += fmt.Sprintf("%s\n", line)
		}
	}

	fmt.Println(string(end[0]))
	for address, change := range changes {
		title := fmt.Sprintf("%s will be <strong>%s</strong>", address, change.action)
		fmt.Println(wrap(title, code(change.details)))
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
