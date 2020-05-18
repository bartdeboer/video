// Copyright 2020 Bart de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var bulkCmd = &cobra.Command{
	Use:   "bulk [path]",
	Args:  cobra.ExactArgs(1),
	Short: "Encode a video",
	Long:  "Encode a video using ffmpeg",
	Run: func(cmd *cobra.Command, args []string) {
		// var ffmpegCmd *exec.Cmd

		dir := strings.Trim(args[0], "\\/")

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {

			// filePath := getSafePath(filepath.Join(dir, file.Name()))
			filePath := filepath.Join(dir, file.Name())
			input := NewVideoFromFile(filePath)
			output := input.NewOutputVideo()
			if input.codec == "h264_cuvid" {
				output.codec = "copy"
			}
			ffmpegCmd := input.getEncodeCommand(output)

			fmt.Println(filePath)

			if initial.DryRun || initial.DetectVolume {
				continue
			}

			ffmpegCmd.Stdout = os.Stdout
			ffmpegCmd.Stderr = os.Stderr

			err := ffmpegCmd.Run()
			if err != nil {
				log.Fatalf("ffmpegCmd.Run() failed with %s\n", err)
			}
		}
	},
}
