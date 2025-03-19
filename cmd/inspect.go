// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/checkpoint-restore/checkpointctl/internal"
	metadata "github.com/checkpoint-restore/checkpointctl/lib"
	"github.com/spf13/cobra"
)

var (
	stats = new(bool)
	mounts = new(bool)
	pID = new(uint32)
	psTree = new(bool)
	psTreeCmd = new(bool)
	psTreeEnv = new(bool)
	files = new(bool)
	sockets = new(bool)
	showAll = new(bool)
	format = new(string)
	showMetdata = new(bool)
	showNetwork = new(bool)
)

func Inspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Display low-level information about a container checkpoint",
		RunE:  inspect,
		Args:  cobra.MinimumNArgs(1),
	}
	flags := cmd.Flags()

	flags.BoolVar(
		stats,
		"stats",
		false,
		"Display checkpoint statistics",
	)
	flags.BoolVar(
		mounts,
		"mounts",
		false,
		"Display an overview of mounts used in the container checkpoint",
	)
	flags.Uint32VarP(
		pID,
		"pid",
		"p",
		0,
		"Display the process tree of a specific PID",
	)
	flags.BoolVar(
		psTree,
		"ps-tree",
		false,
		"Display an overview of processes in the container checkpoint",
	)
	flags.BoolVar(
		psTreeCmd,
		"ps-tree-cmd",
		false,
		"Display an overview of processes in the container checkpoint with full command line arguments",
	)
	flags.BoolVar(
		psTreeEnv,
		"ps-tree-env",
		false,
		"Display an overview of processes in the container checkpoint with their environment variables",
	)
	flags.BoolVar(
		files,
		"files",
		false,
		"Display the open file descriptors for processes in the container checkpoint",
	)
	flags.BoolVar(
		sockets,
		"sockets",
		false,
		"Display the open sockets for processes in the container checkpoint",
	)
	flags.BoolVar(
		showAll,
		"all",
		false,
		"Show all information about container checkpoints",
	)
	flags.StringVar(
		format,
		"format",
		"tree",
		"Specify the output format: tree or json",
	)
	flags.BoolVar(
		showMetdata,
		"metadata",
		false,
		"Show metadata about the container",
	)
	flags.BoolVar(
		showNetwork,
		"network",
		false,
		"Display network information from the checkpoint",
	)

	return cmd
}

func inspect(cmd *cobra.Command, args []string) error {
	if *showAll {
		*stats = true
		*mounts = true
		*psTreeCmd = true
		*psTreeEnv = true
		*files = true
		*sockets = true
		*showMetdata = true
		*showNetwork = true
	}

	requiredFiles := []string{metadata.SpecDumpFile, metadata.ConfigDumpFile}

	if *showNetwork {
		requiredFiles = append(requiredFiles, metadata.NetworkStatusFile)
	}

	if *stats {
		requiredFiles = append(requiredFiles, "stats-dump")
	}

	if *pID != 0 {
		// Enable displaying process tree if the PID filter is passed.
		*psTree = true
	}

	if *files {
		// Enable displaying process tree, even if it is not passed.
		// This is necessary to attach the files to the processes
		// that opened them and display this in the tree.
		*psTree = true
		requiredFiles = append(
			requiredFiles,
			// Unpack files.img, fs-*.img, ids-*.img, fdinfo-*.img
			filepath.Join(metadata.CheckpointDirectory, "files.img"),
			filepath.Join(metadata.CheckpointDirectory, "fs-"),
			filepath.Join(metadata.CheckpointDirectory, "ids-"),
			filepath.Join(metadata.CheckpointDirectory, "fdinfo-"),
		)
	}

	if *sockets {
		// Enable displaying process tree, even if it is not passed.
		// This is necessary to attach the sockets to the processes
		// that opened them and display this in the tree.
		*psTree = true
		requiredFiles = append(
			requiredFiles,
			// Unpack files.img, ids-*.img, fdinfo-*.img
			filepath.Join(metadata.CheckpointDirectory, "files.img"),
			filepath.Join(metadata.CheckpointDirectory, "ids-"),
			filepath.Join(metadata.CheckpointDirectory, "fdinfo-"),
		)
	}

	if *psTreeCmd || *psTreeEnv {
		// Enable displaying process tree when using --ps-tree-cmd or --ps-tree-env.
		*psTree = true
		requiredFiles = append(
			requiredFiles,
			// Unpack pagemap-*.img, pages-*.img, and mm-*.img
			filepath.Join(metadata.CheckpointDirectory, "pagemap-"),
			filepath.Join(metadata.CheckpointDirectory, "pages-"),
			filepath.Join(metadata.CheckpointDirectory, "mm-"),
		)
	}

	if *psTree {
		requiredFiles = append(
			requiredFiles,
			// Unpack pstree.img, core-*.img
			filepath.Join(metadata.CheckpointDirectory, "pstree.img"),
			filepath.Join(metadata.CheckpointDirectory, "core-"),
		)
	}

	tasks, err := internal.CreateTasks(args, requiredFiles)
	if err != nil {
		return err
	}
	defer internal.CleanupTasks(tasks)

	// Display network information if requested
	if *showNetwork {
		for _, task := range tasks {
			networkStatus, file, err := metadata.ReadNetworkStatus(task.Dir)
			if err != nil {
				fmt.Printf("Warning: Failed to read network information from %s: %v\n", file, err)
				continue
			}
			fmt.Printf("\nNetwork Information for %s:\n%s", task.Dir, metadata.FormatNetworkInfo(networkStatus))
		}
	}

	switch *format {
	case "tree":
		return internal.RenderTreeView(tasks)
	case "json":
		return internal.RenderJSONView(tasks)
	default:
		return fmt.Errorf("invalid output format: %s", *format)
	}
}
