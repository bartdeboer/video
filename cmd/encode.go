/*
Copyright © 2020 Bart C.C. de Boer <bart.deboer@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

		if initial.Crop {
			input.detectCrop()
		}

		if initial.DetectVolume {
			input.detectVolume()
		}

		output := input.NewOutputVideo()
		ffmpegCmd = input.getEncodeCommand(output)

		if initial.DryRun || initial.DetectVolume {
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
