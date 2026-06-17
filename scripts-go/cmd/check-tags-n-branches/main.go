package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repositories  map[string]string       `yaml:"repositories"`
	ReleaseGroups map[string]ReleaseGroup `yaml:"release_groups"`
}

type ReleaseGroup struct {
	ReleaseBranches []string `yaml:"release_branches"`
	ReleaseTags     []string `yaml:"release_tags"`
	Repos           []string `yaml:"repos"`
}

func main() {
	var filterRepos, filterBranches, filterTags []string
	branchesOnly := false
	tagsOnly := false

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--repo":
			i++
			if i >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "--repo requires an argument")
				os.Exit(1)
			}
			filterRepos = append(filterRepos, os.Args[i])
		case "--branch":
			i++
			if i >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "--branch requires an argument")
				os.Exit(1)
			}
			filterBranches = append(filterBranches, os.Args[i])
		case "--tag":
			i++
			if i >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "--tag requires an argument")
				os.Exit(1)
			}
			filterTags = append(filterTags, os.Args[i])
		case "--branches-only":
			branchesOnly = true
		case "--tags-only":
			tagsOnly = true
		case "--help", "-h":
			fmt.Println("Usage: check-tags-n-branches [--repo <name>]... [--branch <name>]... [--tag <name>]...")
			fmt.Println("                      [--branches-only] [--tags-only]")
			fmt.Println("\nFilters (can be repeated):")
			fmt.Println("  --repo <name>     check only this repo (short name or full prefix/name)")
			fmt.Println("  --branch <name>   check only this branch")
			fmt.Println("  --tag <name>      check only this tag")
			fmt.Println("  --branches-only   skip tag checks")
			fmt.Println("  --tags-only       skip branch checks")
			fmt.Println("\nAll filters match against what is defined in the config.")
			fmt.Println("With no flags, checks everything across all groups.")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown argument: %s\n", os.Args[i])
			os.Exit(1)
		}
	}

	if branchesOnly && tagsOnly {
		fmt.Fprintln(os.Stderr, "--branches-only and --tags-only are mutually exclusive")
		os.Exit(1)
	}

	repoSet := toSet(filterRepos)
	branchSet := toSet(filterBranches)
	tagSet := toSet(filterTags)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	groupNames := make([]string, 0, len(cfg.ReleaseGroups))
	for name := range cfg.ReleaseGroups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	seen := make(map[string]bool)

	for _, groupName := range groupNames {
		group := cfg.ReleaseGroups[groupName]

		for _, repoRef := range group.Repos {
			name := repoName(resolveURL(repoRef, cfg.Repositories))

			if len(repoSet) > 0 && !repoSet[repoRef] && !repoSet[name] {
				continue
			}
			if seen[repoRef] {
				continue
			}
			seen[repoRef] = true

			var branches, tags []string
			if !tagsOnly {
				branches = filterSlice(group.ReleaseBranches, branchSet)
			}
			if !branchesOnly {
				tags = filterSlice(group.ReleaseTags, tagSet)
			}

			if len(branches) == 0 && len(tags) == 0 {
				continue
			}

			url := resolveURL(repoRef, cfg.Repositories)

			fmt.Printf("=== %s ===\n", name)

			if len(branches) > 0 {
				found, _ := checkRefsExist(url, "refs/heads/", branches)
				printResults("Branches", branches, found)
			}

			if len(tags) > 0 {
				found, _ := checkRefsExist(url, "refs/tags/", tags)
				printResults("Tags", tags, found)
			}

			fmt.Println()
		}
	}
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func filterSlice(items []string, filter map[string]bool) []string {
	if len(filter) == 0 {
		return items
	}
	var result []string
	for _, item := range items {
		if filter[item] {
			result = append(result, item)
		}
	}
	return result
}

func checkRefsExist(url, prefix string, expected []string) (found map[string]bool, missing []string) {
	patterns := make([]string, len(expected))
	for i, ref := range expected {
		patterns[i] = prefix + ref
	}

	args := append([]string{"ls-remote", url}, patterns...)
	out, err := exec.Command("git", args...).Output()

	found = make(map[string]bool)
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			ref := parts[1]
			if strings.HasSuffix(ref, "^{}") {
				continue
			}
			ref = strings.TrimPrefix(ref, prefix)
			found[ref] = true
		}
	}

	for _, ref := range expected {
		if !found[ref] {
			missing = append(missing, ref)
		}
	}
	return found, missing
}

func printResults(label string, expected []string, found map[string]bool) {
	fmt.Printf("  %s:\n", label)
	for _, ref := range expected {
		if found[ref] {
			fmt.Printf("    [ok]      %s\n", ref)
		} else {
			fmt.Printf("    [MISSING] %s\n", ref)
		}
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
