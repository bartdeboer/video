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
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var DetectVolume bool
var DetectOnly bool
var Crop bool
var OutputPath = "D:\\Media\\Movies [Reencoded]\\"
var AudioRate int // k
var AudioCodec string
var AudioChannels int
var FileSize int // MB
var Size string
var Preset string
var Sizes = map[string]int{
	"480p":  720,
	"576p":  720,
	"720p":  1280,
	"1080p": 1920,
	"1440p": 2560,
	"2160p": 3840,
}

type Video struct {
	file          string
	width         int
	height        int
	duration      float64
	rate          int
	codec         string
	audioRate     int
	audioCodec    string
	audioChannels int
	audioLayout   string
	cropWidth     int
	cropHeight    int
	cropX         int
	cropY         int
	cropTop       int
	cropBottom    int
	cropLeft      int
	cropRight     int
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
	keyValues := map[string]string{}
	scanner := bufio.NewScanner(stdout)
	cmd.Start()
	for scanner.Scan() {
		text := scanner.Text()
		key, value := getKeyStringValue(text)
		keyValues[key] = value
		// fmt.Printf("%s\n", text)
	}
	return keyValues, scanner.Err()
}

func (input *Video) initVideo() (int, int) {
	fmt.Print("Get video info\n")
	ffprobCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_format",
		"-show_streams",
		// "-show_entries", "stream=width,height",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd)
	if err != nil {
		log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
	}
	width, _ := strconv.ParseInt(keyValues["width"], 10, 0)
	height, _ := strconv.ParseInt(keyValues["height"], 10, 0)
	duration, _ := strconv.ParseFloat(keyValues["duration"], 64)
	rate, _ := strconv.ParseInt(keyValues["bit_rate"], 10, 0)
	input.width = int(width)
	input.height = int(height)
	input.cropWidth = int(width)
	input.cropHeight = int(height)
	input.duration = duration
	input.codec = keyValues["codec_name"]
	input.rate = int(rate / 1000)
	return int(width), int(height)
}

func (input *Video) initAudio() {
	fmt.Print("Get audio info\n")
	ffprobCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_streams",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd)
	if err != nil {
		log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
	}
	rate, _ := strconv.ParseInt(keyValues["bit_rate"], 10, 0)
	channels, _ := strconv.ParseInt(keyValues["channels"], 10, 0)
	input.audioCodec = keyValues["codec_name"]
	input.audioRate = int(rate / 1000)
	input.audioChannels = int(channels)
	input.audioLayout = keyValues["channel_layout"]
}

func (input *Video) initCropDetect() (int, int, int, int) {
	fmt.Print("Detecting black bars\n")
	ffmpegCmd := exec.Command("ffmpeg",
		"-y", "-hide_banner",
		"-hwaccel", "cuda",
		// "-hwaccel_output_format", "cuda",
		"-c:v", "h264_cuvid",
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
		// fmt.Printf("%q\n", submatches[0])
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

func (output *Video) initOutput(input Video) {
	output.width = input.width
	output.height = input.height
	if outputWidth, ok := Sizes[Size]; ok {
		output.width = outputWidth
		resizeRatio := float64(output.width) / float64(input.width)
		output.height = int(resizeRatio * float64(input.height))
	}
	if input.cropHeight > 0 {
		resizeRatio := float64(output.width) / float64(input.width)
		output.height = int(resizeRatio * float64(input.cropHeight))
		output.cropTop = input.cropY
		output.cropBottom = input.height - (input.cropY + input.cropHeight)
		output.cropLeft = input.cropX
		output.cropRight = input.width - (input.cropX + input.cropWidth)
	}
	// Copy audio by default
	output.audioRate = 0
	output.audioChannels = 0
	output.audioCodec = "copy"
	// If audio rate is specified default to AC3
	if AudioRate > 0 && (input.audioRate == 0 || AudioRate < input.audioRate) {
		output.audioRate = AudioRate
		output.audioCodec = "ac3"
	}
	// If audio channels are specified default to AC3
	if AudioChannels > 0 && AudioChannels != input.audioChannels {
		output.audioChannels = AudioChannels
		output.audioCodec = "ac3"
	}
	// If audio requires encoding for 2 channels default to AAC
	if output.audioCodec != "copy" && output.audioChannels == 2 {
		output.audioCodec = "aac"
	}
	// If codec is specified overrule them all
	if AudioCodec != "" {
		output.audioCodec = AudioCodec
	}
	if FileSize > 0 {
		output.rate = int((float64(FileSize) * 8192 / input.duration) - float64(output.audioRate))
	}
}

func (input *Video) initDetectVolume() /* float64 */ {
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
}

func getEncodeCommand(input Video, output Video) *exec.Cmd {
	var args []string
	args = append(args,
		"-y", "-hide_banner",
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-c:v", "h264_cuvid",
	)
	if input.cropHeight > 0 {
		args = append(args,
			"-crop", strconv.FormatInt(int64(output.cropTop), 10)+
				"x"+strconv.FormatInt(int64(output.cropBottom), 10)+
				"x"+strconv.FormatInt(int64(output.cropLeft), 10)+
				"x"+strconv.FormatInt(int64(output.cropRight), 10),
		)
	}
	if input.width != output.width {
		args = append(args,
			"-resize", strconv.FormatInt(int64(output.width), 10)+
				"x"+strconv.FormatInt(int64(output.height), 10),
		)
	}
	args = append(args, "-i", input.file)
	args = append(args,
		"-c:v", "h264_nvenc",
		"-rc:v", "vbr_hq",
		"-cq:v", "20",
		"-profile:v", "main",
		"-max_muxing_queue_size", "800",
		// "-to", "600",
		// "-af", "pan=stereo|FL < 1.0*FL + 0.707*FC + 0.707*BL|FR < 1.0*FR + 0.707*FC + 0.707*BR",
	)

	if output.rate > 0 {
		args = append(args,
			"-b:v", strconv.FormatInt(int64(output.rate), 10)+"k",
			"-maxrate:v", strconv.FormatInt(int64(output.rate*2), 10)+"k",
		)
	}

	args = append(args, "-c:a", output.audioCodec)

	if output.audioRate > 0 {
		args = append(args, "-b:a", strconv.FormatInt(int64(output.audioRate), 10)+"k")
	}

	if output.audioChannels > 0 {
		args = append(args, "-ac", strconv.FormatInt(int64(output.audioChannels), 10))
	}

	args = append(args, output.file)

	return exec.Command("ffmpeg", args...)
}

var encodeCmd = &cobra.Command{
	Use:   "encode [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Encode a video",
	Long:  "Encode a video using ffmpeg",
	Run: func(cmd *cobra.Command, args []string) {
		if Preset == "telegram" {
			Size = "720p"
			FileSize = 1450
			AudioRate = 128
			AudioChannels = 2
			AudioCodec = "aac"
		}
		var ffmpegCmd *exec.Cmd
		input := Video{}
		output := Video{}
		input.file = args[0]
		output.file = OutputPath + filepath.Base(strings.TrimSuffix(input.file, filepath.Ext(input.file))) + ".720p.mp4"
		if DetectVolume {
			input.initDetectVolume()
		}
		input.initVideo()
		input.initAudio()
		if Crop {
			input.initCropDetect()
		}
		output.initOutput(input)

		ffmpegCmd = getEncodeCommand(input, output)

		fmt.Printf("Input file: %s\n", input.file)
		fmt.Printf("Input duration: %f\n", input.duration)
		fmt.Printf("Input video codec: %s\n", input.codec)
		fmt.Printf("Input width: %d\n", input.width)
		fmt.Printf("Input height: %d\n", input.height)
		fmt.Printf("Input video rate: %d\n", input.rate)
		fmt.Printf("Input audio codec: %s\n", input.audioCodec)
		fmt.Printf("Input audio rate: %d\n", input.audioRate)
		fmt.Printf("Input audio channels: %d\n", input.audioChannels)
		fmt.Printf("Input audio channel layout: %s\n", input.audioLayout)
		fmt.Printf("Output file: %s\n", output.file)
		fmt.Printf("Output width: %d\n", output.width)
		fmt.Printf("Output height: %d\n", output.height)
		fmt.Printf("Output rate: %d\n", output.rate)
		fmt.Printf("Output crop top: %d\n", output.cropTop)
		fmt.Printf("Output crop bottom: %d\n", output.cropBottom)
		fmt.Printf("Output crop left: %d\n", output.cropLeft)
		fmt.Printf("Output crop right: %d\n", output.cropRight)
		fmt.Printf("Output audio codec: %s\n", output.audioCodec)
		fmt.Printf("Output audio rate: %d\n", output.audioRate)
		fmt.Printf("Output audio channels: %d\n", output.audioChannels)
		fmt.Printf("\n%+v\n\n", ffmpegCmd)

		if DetectOnly || DetectVolume {
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

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().BoolVarP(&DetectVolume, "detect-volume", "", false, "Detect volume")
	encodeCmd.Flags().BoolVarP(&DetectOnly, "detect-only", "", false, "Show video info")
	encodeCmd.Flags().BoolVarP(&Crop, "crop", "c", false, "Crop black bars")
	encodeCmd.Flags().IntVarP(&FileSize, "file-size", "f", 0, "Output file size (MB)")
	encodeCmd.Flags().IntVarP(&AudioRate, "audio-rate", "", 0, "Audio rate (k)")
	encodeCmd.Flags().StringVarP(&AudioCodec, "audio-codec", "", "", "Audio codec")
	encodeCmd.Flags().IntVarP(&AudioChannels, "audio-channels", "", 0, "Audio codec")
	encodeCmd.Flags().StringVarP(&Size, "size", "s", "", "Output resolution (480p, 576p, 720p, 1080p, 1440p or 2160p)")
	encodeCmd.Flags().StringVarP(&Preset, "preset", "p", "", "Preset (telegram)")
}
