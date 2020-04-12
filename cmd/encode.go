/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"bufio"
	"os"
	"os/exec"
	"strings"
	"fmt"
	"strconv"
	"path/filepath"
	"github.com/spf13/cobra"
	"log"
)

var OutputPath = "D:\\Media\\Movies [Reencoded]\\"

type Video struct {
	file string
	width int
	height int
	duration float64
	rate int
}

func getKeyIntValue(input string) (string, int) {
	arr := strings.SplitN(string(input), "=", 2)
	key := arr[0]
	value, _ := strconv.ParseInt(arr[1], 10, 0)
	return key, int(value)
}

func (input *Video) getDimensions() (int, int) {
	ffprobCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	stdout, err := ffprobCmd.StdoutPipe()
	if err != nil {
        log.Fatalf("ffprobCmd.StdoutPipe() failed with %s\n", err)
	}
	buf := bufio.NewReader(stdout)
	ffprobCmd.Start()
	widthLine, _, _ := buf.ReadLine()
	heightLine, _, _ := buf.ReadLine()
	_, width := getKeyIntValue(string(widthLine))
	_, height := getKeyIntValue(string(heightLine))
	err = ffprobCmd.Wait()
	if err != nil {
		log.Fatalf("ffprobCmd.Start() failed with %s\n", err)
	}
	return width, height
}

func (input *Video) getDuration() float64 {
	ffprobCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "format=duration",
		"-of", "csv=s=x:p=0",
		"-i", input.file,
	)
    out, err := ffprobCmd.CombinedOutput()
    if err != nil {
        log.Fatalf("ffprobCmd.CombinedOutput() failed with %s\n", err)
	}
	duration, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return duration
}

func getCuvidResizeCommand(input Video, output Video) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-y", "-hide_banner",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-c:v", "h264_cuvid",
		"-resize", strconv.FormatInt(int64(output.width), 10) +
			"x" + strconv.FormatInt(int64(output.height), 10),
		"-i", input.file,
		"-c:v", "h264_nvenc",
		"-rc:v", "vbr_hq",
		"-cq:v", "20",
		"-b:v", strconv.FormatInt(int64(output.rate), 10) + "k",
		"-maxrate:v", "4500k",
		"-profile:v", "main",
		"-max_muxing_queue_size", "800",
		"-c:a", "aac",
		"-b:a", "128k",
		"-af", "pan=stereo|FL < 1.0*FL + 0.707*FC + 0.707*BL|FR < 1.0*FR + 0.707*FC + 0.707*BR",
		output.file,
	)
}

func getNvEncCommand(input Video, output Video) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-y", "-hide_banner",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-c:v", "h264_cuvid",
		"-i", input.file,
		"-c:v", "h264_nvenc",
		"-rc:v", "vbr_hq",
		"-cq:v", "20",
		"-b:v", strconv.FormatInt(int64(output.rate), 10) + "k",
		"-maxrate:v", "4500k",
		"-profile:v", "main",
		"-max_muxing_queue_size", "800",
		"-c:a", "aac",
		"-b:a", "128k",
		"-af", "pan=stereo|FL < 1.0*FL + 0.707*FC + 0.707*BL|FR < 1.0*FR + 0.707*FC + 0.707*BR",
		output.file,
	)
}

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode a video",
	Long: `Encode a video using ffmpeg`,
	Run: func(cmd *cobra.Command, args []string) {
		var ffmpegCmd *exec.Cmd
		input := Video {}
		input.file =args[0]
		input.duration = input.getDuration()
		width, height := input.getDimensions()
		input.width = width
		input.height = height

		output := Video {}
		output.file = OutputPath + filepath.Base(strings.TrimSuffix(input.file, filepath.Ext(input.file))) + ".720p.mp4"
		output.rate = int((1450 * 8192 / input.duration) - 128)
		output.width = 1280
		output.height = int(1280 / float64(input.width) * float64(input.height))

		if (input.width == 1280) {
			ffmpegCmd = getNvEncCommand(input, output)
		} else {
			ffmpegCmd = getCuvidResizeCommand(input, output)
		}

		fmt.Printf("Input file: %s\n", input.file)
		fmt.Printf("Input duration: %f\n", input.duration)
		fmt.Printf("Input width: %d\n", input.width)
		fmt.Printf("Input height: %d\n", input.height)
		fmt.Printf("Output file: %s\n", output.file)
		fmt.Printf("Output width: %d\n", output.width)
		fmt.Printf("Output height: %d\n", output.height)
		fmt.Printf("Output rate: %d\n", output.rate)
		fmt.Printf("\n%+v\n\n", ffmpegCmd)

		ffmpegCmd.Stdout = os.Stdout
		ffmpegCmd.Stderr = os.Stderr

		err := ffmpegCmd.Run()
		if err != nil {
			log.Fatalf("ffmpegCmd.Run() failed with %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	// encodeCmd.Flags().StringVarP(&InputFile, "inputFile", "s", "", "Input file to encode")
}
