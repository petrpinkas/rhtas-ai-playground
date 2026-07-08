package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const localConfigFile = ".projects.yaml"

type Config struct {
	Repositories map[string]string  `yaml:"repositories"`
	Projects     map[string]Project `yaml:"projects"`
}

type Project struct {
	DefaultBranch string `yaml:"default_branch"`
	Repos         []Repo `yaml:"repos"`
}

type Repo struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch,omitempty"`
}

type LocalConfig struct {
	Repositories  map[string]string `yaml:"repositories"`
	DefaultBranch string            `yaml:"default_branch"`
	Repos         []Repo            `yaml:"repos"`
}

func main() {
	callerDir := os.Getenv("GET_PROJECTS_CWD")
	if callerDir == "" {
		callerDir, _ = os.Getwd()
	}

	forceCleanup := false
	args := []string{}
	for i := 1; i < len(os.Args); i++ {
		a := os.Args[i]
		if a == "--force-cleanup" {
			forceCleanup = true
		} else {
			args = append(args, a)
		}
	}

	if len(args) >= 1 && args[0] == "init" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: get-projects init <project> [<project2> ...]\n")
			os.Exit(1)
		}
		if err := runInit(args[1:], callerDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  get-projects init <project> [<project2> ...]  Initialize .projects.yaml in current directory\n")
		fmt.Fprintf(os.Stderr, "  get-projects [--force-cleanup] Update repos from .projects.yaml\n")
		os.Exit(1)
	}

	if err := runUpdate(callerDir, forceCleanup); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInit(projectNames []string, callerDir string) error {
	configPath := filepath.Join(callerDir, localConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%s already exists; delete it first to re-initialize", localConfigFile)
	}

	cfg, err := loadBundledConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	usedPrefixes := map[string]string{}
	seen := map[string]string{}        // url -> resolved branch
	seenSource := map[string]string{}  // url -> project name that added it
	var mergedRepos []Repo
	defaultBranch := ""

	for i, projectName := range projectNames {
		project, ok := cfg.Projects[projectName]
		if !ok {
			available := []string{}
			for name := range cfg.Projects {
				available = append(available, name)
			}
			return fmt.Errorf("unknown project %q; available: %s", projectName, strings.Join(available, ", "))
		}

		if i == 0 {
			defaultBranch = project.DefaultBranch
		}

		for _, repo := range project.Repos {
			branch := repo.Branch
			if branch == "" {
				branch = project.DefaultBranch
			}

			if prev, exists := seen[repo.URL]; exists {
				if prev != branch {
					return fmt.Errorf("conflict for %s: branch %q (from %s) vs %q (from %s)",
						repo.URL, prev, seenSource[repo.URL], branch, projectName)
				}
				continue
			}

			seen[repo.URL] = branch
			seenSource[repo.URL] = projectName

			outRepo := Repo{URL: repo.URL}
			if branch != defaultBranch {
				outRepo.Branch = branch
			}
			mergedRepos = append(mergedRepos, outRepo)

			parts := strings.SplitN(repo.URL, "/", 2)
			if len(parts) == 2 {
				if prefix, ok := cfg.Repositories[parts[0]]; ok {
					usedPrefixes[parts[0]] = prefix
				}
			}
		}
	}

	local := LocalConfig{
		Repositories:  usedPrefixes,
		DefaultBranch: defaultBranch,
		Repos:         mergedRepos,
	}

	data, err := yaml.Marshal(&local)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", localConfigFile, err)
	}

	fmt.Printf("Created %s with %d repositories (from projects: %s)\n",
		configPath, len(mergedRepos), strings.Join(projectNames, ", "))
	return nil
}

func runUpdate(callerDir string, forceCleanup bool) error {
	configPath := filepath.Join(callerDir, localConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found; run 'get-projects init <project>' first", localConfigFile)
		}
		return fmt.Errorf("reading %s: %w", localConfigFile, err)
	}

	var local LocalConfig
	if err := yaml.Unmarshal(data, &local); err != nil {
		return fmt.Errorf("parsing %s: %w", localConfigFile, err)
	}

	baseDir := callerDir

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
	for _, repo := range local.Repos {
		fullURL := resolveURL(repo.URL, local.Repositories)
		name := repoName(fullURL)
		branch := repo.Branch
		if branch == "" {
			branch = local.DefaultBranch
		}
		if branch == "" {
			return fmt.Errorf("no branch configured for %q and no default_branch set", repo.URL)
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
			if err := git(dir, "fetch", "--all", "--prune"); err != nil {
				fmt.Printf("  WARNING: fetch failed: %v\n", err)
			}
		} else {
			fmt.Printf("--- %-*s  %-*s  [clone] ---\n", maxURL, url, maxBranch, branch)
			if err := git(".", "clone", url, dir); err != nil {
				fmt.Printf("  ERROR: clone failed: %v\n", err)
				cloneFailed = append(cloneFailed, name)
				continue
			}
		}

		if forceCleanup {
			_ = git(dir, "checkout", "--", ".")
			_ = git(dir, "clean", "-fd")
		}

		if branchExists(dir, "origin/"+branch) {
			if err := git(dir, "checkout", branch); err != nil {
				_ = git(dir, "checkout", "-b", branch, "origin/"+branch)
			}
			_ = git(dir, "pull", "--ff-only")
		} else if branchExists(dir, branch) {
			_ = git(dir, "checkout", branch)
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
	for _, ri := range infos {
		dir := ri.dir
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
			fmt.Printf("\n--- %s (on %s) ---\n", ri.name, current)
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

	knownRepos := map[string]bool{}
	for _, ri := range infos {
		knownRepos[ri.name] = true
	}

	entries, _ := os.ReadDir(baseDir)
	var unmanaged []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if knownRepos[name] {
			continue
		}
		if isGitRepo(filepath.Join(baseDir, name)) {
			unmanaged = append(unmanaged, name)
		}
	}
	if len(unmanaged) > 0 {
		fmt.Println()
		fmt.Println("========================================")
		fmt.Println("  Unmanaged repositories")
		fmt.Println("========================================")
		fmt.Printf("\nNot in %s: %s\n", localConfigFile, strings.Join(unmanaged, ", "))
	}

	return nil
}

func loadBundledConfig() (*Config, error) {
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
