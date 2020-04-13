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
	"runtime"
	"regexp"
	"math"
)

var DoVolumeDetect bool
var DoCrop bool
var OutputPath = "D:\\Media\\Movies [Reencoded]\\"
var AudioRate = 128
var OutputSize = 1450 // Mb
var OutputWidth int

type Video struct {
	file string
	width int
	height int
	duration float64
	rate int
	cropWidth int
	cropHeight int
	cropX int
	cropY int
	cropTop int
	cropBottom int
	cropLeft int
	cropRight int
}


func getNullDevice() string {
	if runtime.GOOS == "windows" {
		return "NUL"
	}
	return "/dev/null"
}


func getKeyStringValue(input string) (string, string) {
	arr := strings.SplitN(string(input), "=", 2)
	return arr[0], arr[1]
}


func getKeyIntValue(input string) (string, int, error) {
	arr := strings.SplitN(string(input), "=", 2)
	key := arr[0]
	value, err := strconv.ParseInt(arr[1], 10, 0)
	return key, int(value), err
}


func getKeyValuesFromCommand(cmd *exec.Cmd) (map[string]string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
        log.Fatalf("cmd.StdoutPipe() failed with %s\n", err)
	}
	keyValues := map[string]string {}
    scanner := bufio.NewScanner(stdout)
	cmd.Start()
	for scanner.Scan() {
		key, value := getKeyStringValue(scanner.Text())
		keyValues[key] = value;
	}
	return keyValues, scanner.Err()
}


func (input *Video) initDimensions() (int, int) {
	fmt.Print("Get dimensions\n")
	ffprobCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd)
	if err != nil {
        log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
    }
	width, _ := strconv.ParseInt(keyValues["width"], 10, 0)
	height, _ := strconv.ParseInt(keyValues["height"], 10, 0)
	input.width = int(width)
	input.height = int(height)
	input.cropWidth = int(width)
	input.cropHeight = int(height)
	return int(width), int(height)
}


func (input *Video) initDuration() float64 {
	fmt.Print("Get duration\n")
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
	input.duration = duration
	return duration
}


func (input *Video) initCropDetect() (int, int, int, int) {
	fmt.Print("Detecting black bars\n")
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", input.file,
		"-vf", "fps=1/60,cropdetect=24:16:0",
		"-to", "600",
		"-an",
		"-f", "null",
		getNullDevice(),
	)
	out, _ := ffmpegCmd.CombinedOutput()
	// fmt.Printf("Crop Detect: %s\n", string(out))
	r, _ := regexp.Compile("crop=([0-9]+):([0-9]+):([0-9]+):([0-9]+)")
	matches := r.FindAllStringSubmatch(string(out), -1)
	width, height, x, y := 0, 0, input.width, input.height
	for _, submatches := range matches {
		fmt.Printf("%q\n", submatches[0])
		subWidth, _ := strconv.ParseInt(submatches[1], 10, 0)
		subHeight, _ := strconv.ParseInt(submatches[2], 10, 0)
		subX, _ := strconv.ParseInt(submatches[3], 10, 0)
		subY, _ := strconv.ParseInt(submatches[4], 10, 0)
		input.cropWidth = int(math.Max(float64(width), float64(subWidth)))
		input.cropHeight = int(math.Max(float64(height), float64(subHeight)))
		input.cropX = int(math.Min(float64(x), float64(subX)))
		input.cropY = int(math.Min(float64(y), float64(subY)))
	}
	return input.cropWidth, input.cropHeight, input.cropX, input.cropY
}


func (output *Video) setDimensions(input Video) {
	output.width = OutputWidth
	resizeRatio := float64(OutputWidth) / float64(input.width)
	output.height = int(resizeRatio * float64(input.height))
	if (input.cropHeight > 0) {
		output.height = int(resizeRatio * float64(input.cropHeight))
		output.cropTop = input.cropY
		output.cropBottom = input.height - (input.cropY + input.cropHeight)
		output.cropLeft = input.cropX
		output.cropRight = input.width - (input.cropX + input.cropWidth)
	}
}


func (input *Video) initVolumeDetect() /* float64 */ {
	fmt.Print("Detecting volume levels\n")
	ffmpegCmd := exec.Command("ffmpeg",
		"-hide_banner",
		"-i", input.file,
		"-to", "400",
		"-vn",
		"-filter:a", "volumedetect",
		"-f", "null",
		getNullDevice(),
	)
	out, _ := ffmpegCmd.CombinedOutput()
	r, _ := regexp.Compile("max_volume:[^\\n]+")
	fmt.Println(r.FindString(string(out)))

	// ffmpegCmd.Stdout = os.Stdout
	// ffmpegCmd.Stderr = os.Stderr
	// ffmpegCmd.Run()

    // out, err := ffmpegCmd.CombinedOutput()
	// fmt.Printf("Volume: %s\n", string(out))
    // if err != nil {
    //     log.Fatalf("HERE ffmpegCmd.CombinedOutput() failed with %s\n", err)
	// }
	// volume, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	// return volume
}

func getCuvidResizeCommand(input Video, output Video) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-y", "-hide_banner",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-c:v", "h264_cuvid",
		"-crop", strconv.FormatInt(int64(output.cropTop), 10) +
			"x" + strconv.FormatInt(int64(output.cropBottom), 10) +
			"x" + strconv.FormatInt(int64(output.cropLeft), 10) +
			"x" + strconv.FormatInt(int64(output.cropRight), 10),
		"-resize", strconv.FormatInt(int64(output.width), 10) +
		 	"x" + strconv.FormatInt(int64(output.height), 10),
		"-i", input.file,
		// "-to", "600",
		"-c:v", "h264_nvenc",
		"-rc:v", "vbr_hq",
		"-cq:v", "20",
		"-b:v", strconv.FormatInt(int64(output.rate), 10) + "k",
		"-maxrate:v", "4500k",
		"-profile:v", "main",
		"-max_muxing_queue_size", "800",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		// "-af", "pan=stereo|FL < 1.0*FL + 0.707*FC + 0.707*BL|FR < 1.0*FR + 0.707*FC + 0.707*BR",
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
		"-b:a", strconv.FormatInt(int64(AudioRate), 10) + "k",
		"-ac", "2",
		// "-af", "pan=stereo|FL < 1.0*FL + 0.707*FC + 0.707*BL|FR < 1.0*FR + 0.707*FC + 0.707*BR",
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
		if (DoVolumeDetect) {
			input.initVolumeDetect()
			os.Exit(0)
		}
		input.initDuration()
		input.initDimensions()
		if (DoCrop) {
			input.initCropDetect()
		}

		output := Video {}
		output.file = OutputPath + filepath.Base(strings.TrimSuffix(input.file, filepath.Ext(input.file))) + ".720p.mp4"
		output.rate = int((float64(OutputSize) * 8192 / input.duration) - float64(AudioRate))

		output.setDimensions(input)

		if (input.width == output.width) {
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
		fmt.Printf("Output crop top: %d\n", output.cropTop)
		fmt.Printf("Output crop bottom: %d\n", output.cropBottom)
		fmt.Printf("Output crop left: %d\n", output.cropLeft)
		fmt.Printf("Output crop right: %d\n", output.cropRight)
		fmt.Printf("\n%+v\n\n", ffmpegCmd)

		// os.Exit(0)

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
	encodeCmd.Flags().BoolVarP(&DoVolumeDetect, "volume-detect", "", false, "Detect volume")
	encodeCmd.Flags().BoolVarP(&DoCrop, "crop", "", false, "Crop")
	encodeCmd.Flags().IntVarP(&OutputWidth, "output-width", "", 1280, "Output width")
	// encodeCmd.Flags().StringVarP(&VolumeDetect, "volume-detect", "vd", "", "Detect Volume")
	// encodeCmd.Flags().StringVarP(&InputFile, "inputFile", "s", "", "Input file to encode")
}
