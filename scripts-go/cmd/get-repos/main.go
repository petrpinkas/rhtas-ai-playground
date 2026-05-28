package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repositories map[string]string  `yaml:"repositories"`
	Projects     map[string]Project `yaml:"projects"`
}

type Project struct {
	Dir           string `yaml:"dir"`
	DefaultBranch string `yaml:"default_branch"`
	Repos         []Repo `yaml:"repos"`
}

type Repo struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

func main() {
	forceCleanup := false
	outputDir := ""
	args := []string{}
	for i := 1; i < len(os.Args); i++ {
		a := os.Args[i]
		if a == "--force-cleanup" {
			forceCleanup = true
		} else if a == "--output-dir" {
			i++
			if i >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "--output-dir requires a path argument\n")
				os.Exit(1)
			}
			outputDir = os.Args[i]
		} else {
			args = append(args, a)
		}
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: get-repos <project> [--output-dir <path>] [--force-cleanup]\n")
		os.Exit(1)
	}

	projectName := args[0]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	project, ok := cfg.Projects[projectName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown project %q. Available projects:", projectName)
		for name := range cfg.Projects {
			fmt.Fprintf(os.Stderr, " %s", name)
		}
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	baseDir := outputDir
	if baseDir == "" {
		baseDir = project.Dir
	}
	if baseDir == "" {
		fmt.Fprintf(os.Stderr, "Error: no dir configured for project %q and --output-dir not provided\n", projectName)
		os.Exit(1)
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	type repoInfo struct {
		repo    Repo
		name    string
		fullURL string
		branch  string
		dir     string
	}
	var infos []repoInfo
	maxURL := 0
	maxBranch := 0
	for _, repo := range project.Repos {
		fullURL := resolveURL(repo.URL, cfg.Repositories)
		name := repoName(fullURL)
		branch, err := resolveBranch(repo, project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		dir := filepath.Join(baseDir, name)
		infos = append(infos, repoInfo{repo, name, fullURL, branch, dir})
		if len(fullURL) > maxURL {
			maxURL = len(fullURL)
		}
		if len(branch) > maxBranch {
			maxBranch = len(branch)
		}
	}

	var cloneFailed []string
	var checkoutFailed []string

	for _, ri := range infos {
		name := ri.name
		branch := ri.branch
		url := ri.fullURL
		dir := ri.dir

		if isGitRepo(dir) {
			fmt.Printf("--- %-*s  %-*s  [fetch] ---\n", maxURL, url, maxBranch, branch)
			if err := git(dir, "fetch", "--all", "--prune", "--quiet"); err != nil {
				fmt.Printf("  WARNING: fetch failed: %v\n", err)
			}
		} else {
			fmt.Printf("--- %-*s  %-*s  [clone] ---\n", maxURL, url, maxBranch, branch)
			if err := git(".", "clone", "--quiet", url, dir); err != nil {
				fmt.Printf("  ERROR: clone failed: %v\n", err)
				cloneFailed = append(cloneFailed, name)
				continue
			}
		}

		if forceCleanup {
			_ = git(dir, "checkout", "--", ".")
			_ = git(dir, "clean", "-fd", "--quiet")
		}

		if branchExists(dir, "origin/"+branch) {
			if err := git(dir, "checkout", branch, "--quiet"); err != nil {
				_ = git(dir, "checkout", "-b", branch, "origin/"+branch, "--quiet")
			}
			_ = git(dir, "pull", "--ff-only", "--quiet")
		} else if branchExists(dir, branch) {
			_ = git(dir, "checkout", branch, "--quiet")
		} else {
			checkoutFailed = append(checkoutFailed, name)
			fmt.Printf("  WARNING: branch %q not found\n", branch)
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Local changes report")
	fmt.Println("========================================")

	anyChanges := false
	for _, repo := range project.Repos {
		name := repoName(repo.URL)
		dir := filepath.Join(baseDir, name)
		if !isGitRepo(dir) {
			continue
		}
		status := gitOutput(dir, "status", "--short")
		if status != "" {
			anyChanges = true
			current := gitOutput(dir, "branch", "--show-current")
			if current == "" {
				current = "detached"
			}
			fmt.Printf("\n--- %s (on %s) ---\n", name, current)
			fmt.Println(status)
		}
	}

	if !anyChanges {
		fmt.Println("\nNo local changes in any repository.")
	}

	if len(cloneFailed) > 0 {
		fmt.Printf("\nFailed to clone: %s\n", strings.Join(cloneFailed, ", "))
	}
	if len(checkoutFailed) > 0 {
		fmt.Printf("\nBranch not found in: %s\n", strings.Join(checkoutFailed, ", "))
	}
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config/repositories.yaml")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func resolveURL(raw string, repositories map[string]string) string {
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) == 2 {
		if prefix, ok := repositories[parts[0]]; ok {
			return prefix + "/" + parts[1] + ".git"
		}
	}
	return raw
}

func repoName(url string) string {
	base := filepath.Base(url)
	return strings.TrimSuffix(base, ".git")
}

func resolveBranch(repo Repo, project Project) (string, error) {
	if repo.Branch != "" {
		return repo.Branch, nil
	}
	if project.DefaultBranch != "" {
		return project.DefaultBranch, nil
	}
	return "", fmt.Errorf("no branch configured for %q and project has no default_branch", repo.URL)
}

func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "." {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func branchExists(dir, ref string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = dir
	return cmd.Run() == nil
}
