// Copyright 2020 Bart de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Encode a video",
	Long:  "Encode a video using ffmpeg",
	Run: func(cmd *cobra.Command, args []string) {
		var ffmpegCmd *exec.Cmd
		input := NewVideoFromFile(args[0])

		input.detectVideo(initial.VideoStream)
		input.detectAudio(initial.AudioStream)

		if initial.Crop {
			input.detectCrop()
		}

		if initial.DetectVolume {
			input.detectVolume()
		}

		if initial.InputCodec != "" {
			input.codec = initial.InputCodec
		}

		output := input.NewOutputVideoFromCmdAgrs()
		ffmpegCmd = input.getEncodeCommand(output)

		if initial.DryRun {
			os.Exit(0)
		}

		ffmpegCmd.Stdout = os.Stdout
		ffmpegCmd.Stderr = os.Stderr

		err := ffmpegCmd.Run()
		if err != nil {
			log.Fatalf("ffmpegCmd.Run() failed with %s\n", err)
		}
	},
}
