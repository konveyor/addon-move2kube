package main

import (
	"fmt"
	"os"
	pathlib "path"
	"strings"

	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

// addTags ensure tags created and associated with application.
// Ensure tag exists and associated with the application.
func addTags(application *api.Application, names ...string) error {
	addon.Activity("Adding tags: %v", names)
	appTags := appTags(application)
	// Fetch tags and tag types.
	tpMap, err := tpMap()
	if err != nil {
		return err
	}
	tagMap, err := tagMap()
	if err != nil {
		return err
	}
	// Ensure type exists.
	wanted := api.TagType{
		Name:  "DIRECTORY",
		Color: "#2b9af3",
		Rank:  3,
	}
	tp, found := tpMap[wanted.Name]
	if !found {
		tp = wanted
		if err := addon.TagType.Create(&tp); err != nil {
			return err
		}
		tpMap[tp.Name] = tp
	} else {
		if wanted.Rank != tp.Rank || wanted.Color != tp.Color {
			return &hub.SoftError{Reason: "Tag (TYPE) conflict detected."}
		}
	}
	// Add tags.
	for _, name := range names {
		if _, found := appTags[name]; found {
			continue
		}
		wanted := api.Tag{
			Name:    name,
			TagType: api.Ref{ID: tp.ID},
		}
		tg, found := tagMap[wanted.Name]
		if !found {
			tg = wanted
			if err := addon.Tag.Create(&tg); err != nil {
				return err
			}
			tagMap[wanted.Name] = tg
		} else {
			if wanted.TagType.ID != tg.TagType.ID {
				return &hub.SoftError{Reason: "Tag conflict detected."}
			}
		}
		addon.Activity("[TAG] Associated: %s.", tg.Name)
		application.Tags = append(
			application.Tags,
			api.Ref{ID: tg.ID},
		)
	}
	// Update application.
	return addon.Application.Update(application)
}

// tagMap builds a map of tags by name.
func tagMap() (map[string]api.Tag, error) {
	list, err := addon.Tag.List()
	if err != nil {
		return nil, err
	}
	m := map[string]api.Tag{}
	for _, tag := range list {
		m[tag.Name] = tag
	}
	return m, nil
}

// tpMap builds a map of tag types by name.
func tpMap() (map[string]api.TagType, error) {
	list, err := addon.TagType.List()
	if err != nil {
		return nil, err
	}
	m := map[string]api.TagType{}
	for _, t := range list {
		m[t.Name] = t
	}
	return m, nil
}

// appTags builds map of associated tags.
func appTags(application *api.Application) map[string]uint {
	m := map[string]uint{}
	for _, ref := range application.Tags {
		m[ref.Name] = ref.ID
	}
	return m
}

// commitResources commits the resources to the Git repo.
// func commitResources(SourceDir, groupId, artifactId string) error {
func commitResources(
	repo repository.Repository,
	repoDir,
	inputBranch,
	inputDir,
	outputBranch,
	outputDir,
	transformOutputDir string,
	skipConfig bool,
	commitMessage string,
) error {
	if outputBranch == "" {
		outputBranch = "move2kube-output"
	}
	if outputDir == "" {
		outputDir = pathlib.Join(inputDir, "move2kube-output")
	} else {
		outputDir = pathlib.Join(repoDir, outputDir)
	}
	if commitMessage == "" {
		commitMessage = "feat: add move2kube transform output"
	}
	if inputBranch == "" {
		addon.Activity("Trying to detect the current branch.")
		t1, err := getCurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get the current git branch. Error: %w", err)
		}
		inputBranch = t1
	}
	addon.Activity("The current branch is '%s'", inputBranch)

	// Create a new branch to store the output.
	if err := repo.Branch(outputBranch); err != nil {
		return fmt.Errorf("failed to switch to a new branch. Error: %w", err)
	}

	// Copy the output into the repo.
	if err := os.MkdirAll(outputDir, 0775); err != nil {
		return fmt.Errorf("failed to create the output directory at path '%s'. Error: %w", outputDir, err)
	}
	filesToCopy := []string{"m2k-graph.json", "m2kconfig.yaml", "m2kqacache.yaml", transformOutputDir}
	if skipConfig {
		filesToCopy = []string{transformOutputDir}
	}
	cmd := command.Command{Path: "/usr/bin/cp"}
	cmd.Options.Add("-r", filesToCopy...)
	cmd.Options.Add(outputDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy the output to the repo directory. Error: %w", err)
	}

	// Commit and push the output.
	if err := repo.Commit([]string{"-A"}, commitMessage); err != nil {
		return fmt.Errorf("failed to commit and push all the files. Error: %w", err)
	}

	// Checkout the original branch for future runs.
	if err := repo.Branch(inputBranch); err != nil {
		return fmt.Errorf("failed to switch back to the original branch. Error: %w", err)
	}

	return nil
}

func getCurrentBranch() (string, error) {
	cmd := command.Command{
		Path:    "/usr/bin/git",
		Options: []string{"symbolic-ref", "HEAD", "2>/dev/null"},
	}
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run the git symbolic-ref command. Error: %w", err)
	}
	repoBranchHead := string(cmd.Output)
	return strings.TrimPrefix(repoBranchHead, "refs/heads/"), nil
}
