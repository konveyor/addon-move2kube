/*
Move2Kube addon adapter.
This is an addon adapter that clones a repo and runs Move2Kube transform
on the source code inside that repo. Output is written to a bucket.
*/
package main

import (
	"fmt"
	"os"
	pathlib "path"

	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
)

var (
	// hub integration.
	addon = hub.Addon
	Log   = hub.Log
)

func main() {
	addon.Run(func() error {
		addon.Activity("Fetching the application.")
		application, err := addon.Task.Application()
		if err != nil {
			return fmt.Errorf("failed to get the application. Error: %w", err)
		}

		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get the current working directory. Error: %w", err)
		}

		addon.Activity("Starting the SSH agent.")
		agent := ssh.Agent{}
		if err := agent.Start(); err != nil {
			return fmt.Errorf("failed to start the SSH agent. Error: %w", err)
		}

		// Task update: Update the task with total number of
		// items to be processed by the addon.
		addon.Total(3)

		addon.Activity("Fetching the input from the repository.")
		inputDir := pathlib.Join(pwd, "input")
		repo, err := repository.New(inputDir, application)
		if err != nil {
			return fmt.Errorf("failed to create a new repository object. Error: %w", err)
		}
		if err := repo.Fetch(); err != nil {
			return fmt.Errorf("failed to fetch data from the repository. Error: %w", err)
		}
		addon.Increment()
		addon.Activity("Fetched the input from the repo.")

		addon.Activity("Running Move2Kube transform on the input directory.")
		outputDir := pathlib.Join(application.Bucket, "output")
		if err := runMove2Kube(inputDir, outputDir); err != nil {
			return fmt.Errorf("failed to run Move2Kube transform. Error: %w", err)
		}
		addon.Increment()
		addon.Activity("Transformation finished.")

		if err := commitResources(repo, inputDir, outputDir); err != nil {
			return fmt.Errorf("failed to commit the Move2Kube transform output. Error: %w", err)
		}
		addon.Increment()
		addon.Activity("Pushed the output to the repo.")

		// Task update: update the current addon activity.
		// Set facts.
		application.Facts["Transformed"] = true
		if err := addon.Application.Update(application); err != nil {
			return fmt.Errorf("failed to update the application. Error: %w", err)
		}
		// Add tags.
		if err := addTags(application, "TRANSFORMED"); err != nil {
			return fmt.Errorf("failed to add tags to the application. Error: %w", err)
		}
		return nil
	})
}
