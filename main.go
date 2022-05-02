package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

type Plan struct {
	changes map[string]*ResourceChange
	summary string
}

type ResourceChange struct {
	action  string
	details string
}

var (
	segmentPattern *regexp.Regexp = regexp.MustCompile(`^.*# (.*) will be (.*)$`)
	endPattern     *regexp.Regexp = regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.`)

	planFile          string
	githubToken       string
	pullRequestNumber int
)

func init() {
	flag.StringVar(&planFile, "plan-file", "", "Plan file")
	flag.StringVar(&githubToken, "github-token", "", "Github token")
	flag.IntVar(&pullRequestNumber, "pull-request-number", -1, "Pull Request number")
}

func main() {
	flag.Parse()

	githubRepo := strings.SplitN(os.Getenv("GITHUB_REPOSITORY"), "/", 2)
	owner, repo := githubRepo[0], githubRepo[1]

	ctx := context.Background()
	fmt.Printf("Token %d\n", len(githubToken))
	githubClient := setupGitHubClient(ctx, githubToken)
	pullRequestsComments, _, err := githubClient.PullRequests.ListComments(ctx, owner, repo, pullRequestNumber, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, comment := range pullRequestsComments {
		fmt.Printf("comment: %+v:", comment)
	}

	plan, err := formatPlan(os.Stdin)
	if err != nil {
		log.Fatalf("format plan: %v", err)
	}

	fmt.Println(plan.summary)
	for address, change := range plan.changes {
		title := fmt.Sprintf("%s will be <strong>%s</strong>", address, change.action)
		fmt.Println(wrap(title, code(change.details)))
	}
}

func setupGitHubClient(ctx context.Context, token string) *github.Client {
	if len(token) == 0 {
		log.Fatal("Too short")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func formatPlan(in io.Reader) (*Plan, error) {
	changes := make(map[string]*ResourceChange)
	resourceAddress := ""
	var planSummary [][]byte

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Bytes()

		planSummary = endPattern.FindSubmatch(line)
		if planSummary != nil {
			if len(planSummary) != 4 {
				return nil, fmt.Errorf("invalid plan summary: %s", line)
			}
			break
		}

		segment := segmentPattern.FindSubmatch(line)
		if segment != nil {
			if len(segment) != 3 {
				return nil, fmt.Errorf("invalid segment separator: %s", line)
			}
			resourceAddress = string(segment[1])
			changes[resourceAddress] = &ResourceChange{
				action: string(segment[2]),
			}
		}

		if resourceAddress != "" {
			changes[resourceAddress].details += fmt.Sprintf("%s\n", line)
		}
	}

	return &Plan{
		changes: changes,
		summary: string(planSummary[1]),
	}, nil
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
