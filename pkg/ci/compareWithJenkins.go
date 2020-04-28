package ci

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const ORGANISATION = "alphagov"
const JENKINS_CONTEXT = "continuous-integration/jenkins/pr-head"
const CONCOURSE_CONTEXT = "concourse-ci/status"

var githubClient *github.Client

type PrResult struct {
	repoName string
	details  *github.PullRequest
	statuses []*github.RepoStatus
}

func CompareWithJenkins(repoOption string, numberOfPrs int) error {
	initClient()

	repos, err := parseRepoOption(repoOption)
	if err != nil {
		return err
	}

	prsAndStatuses, err := getPrsAndStatuses(repos, numberOfPrs)
	if err != nil {
		return err
	}

	printSummary(prsAndStatuses)

	return nil
}

func parseRepoOption(repoOption string) ([]string, error) {
	if repoOption == "all" {
		repos, err := getPayRepos()
		if err != nil {
			return nil, err
		}
		return repos, nil
	} else {
		return []string{repoOption}, nil
	}
}

func printSummary(prsAndStatuses []PrResult) {
	output := false
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Repo Name\tPR Title\tState\tJenkins Status\tConcourse Status")
	for _, pr := range prsAndStatuses {
		concourseStatus := getLastStatusWithContextOf(pr.statuses, CONCOURSE_CONTEXT)
		jenkinsStatus := getLastStatusWithContextOf(pr.statuses, JENKINS_CONTEXT)
		if formatState(jenkinsStatus) != formatState(concourseStatus) {
			output = true
			line := fmt.Sprintf("%s\t%.15s...\t%s\t%s\t%s",
				pr.repoName,
				*pr.details.Title,
				*pr.details.State,
				formatState(jenkinsStatus),
				formatState(concourseStatus))

			fmt.Fprintln(w, line)
		}
	}
	if !output {
		fmt.Println("All states match across the repo.")
	} else {
		w.Flush()
	}
}

func getLastStatusWithContextOf(statuses []*github.RepoStatus, context string) string {
	for _, status := range statuses {
		if *status.Context == context {
			return *status.State
		}
	}
	return ""
}

func formatState(state string) string {
	if state == "error" {
		return "failure"
	}

	if state == "" {
		return "no build"
	}

	return state
}

func initClient() {
	var tokenClient *http.Client
	oauthToken := os.Getenv("PAY_CLI_GITHUB_ACCESS_TOKEN")
	if oauthToken != "" {
		ctx := context.Background()
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: oauthToken},
		)
		tokenClient = oauth2.NewClient(ctx, tokenSource)
	}
	githubClient = github.NewClient(tokenClient)
}

func getPrsAndStatuses(repos []string, numberOfPrs int) ([]PrResult, error) {
	var results []PrResult
	for _, repo := range repos {
		prs, err := getPullRequests(repo, numberOfPrs)
		if err != nil {
			return results, err
		}

		for _, pr := range prs {
			statuses, err := getStatusesForPr(*pr)
			if err != nil {
				return results, err
			}
			results = append(results, PrResult{repo, pr, statuses})
		}
	}

	return results, nil
}

func getPullRequests(repo string, numberOfPrs int) ([]*github.PullRequest, error) {
	var pullRequests []*github.PullRequest

	// for loop deals with pagination
	for {
		opts := &github.PullRequestListOptions{}
		prs, resp, err := githubClient.PullRequests.List(context.Background(), ORGANISATION, repo, opts)

		if err != nil {
			return nil, err
		}

		pullRequests = append(pullRequests, prs...)

		if resp.NextPage == 0 || len(pullRequests) >= numberOfPrs {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(pullRequests) > numberOfPrs {
		pullRequests = pullRequests[:numberOfPrs]
	}

	return prDateFilter(pullRequests), nil
}

func prDateFilter(prs []*github.PullRequest) (ret []*github.PullRequest) {
	whenConcourseWasEnabled := time.Date(2020, 4, 27, 0, 0, 0, 0, time.UTC)
	for _, pr := range prs {
		if pr.UpdatedAt.After(whenConcourseWasEnabled) {
			ret = append(ret, pr)
		}
	}
	return
}

func getStatusesForPr(pr github.PullRequest) ([]*github.RepoStatus, error) {
	var statuses []*github.RepoStatus
	opts := &github.ListOptions{
		PerPage: 10,
	}

	for {
		reposStatus, resp, err := githubClient.Repositories.ListStatuses(
			context.Background(),
			ORGANISATION,
			*pr.GetHead().Repo.Name,
			*pr.GetHead().SHA, opts)

		if err != nil {
			return nil, err
		}

		statuses = append(statuses, reposStatus...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].UpdatedAt.After(*statuses[j].UpdatedAt)
	})

	return statuses, nil
}

func getPayRepos() ([]string, error) {
	// could get these by calling github but the module only takes an org
	// and there a lot of non-pay repos to retreive before we could filter them.
	// Revisit if necessary.
	return []string{
		"pay-adminusers",
		"pay-cardid",
		"pay-cli",
		"pay-connector",
		"pay-direct-debit-connector",
		"pay-direct-debit-frontend",
		"pay-endtoend",
		"pay-frontend",
		"pay-java-commons",
		"pay-js-commons",
		"pay-ledger",
		"pay-notifications",
		"pay-omnibus",
		"pay-product-page",
		"pay-products",
		"pay-products-ui",
		"pay-publicapi",
		"pay-publicauth",
		"pay-selfservice",
		"pay-toolbox"}, nil
}
