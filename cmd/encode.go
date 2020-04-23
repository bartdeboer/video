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
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var DetectVolume bool
var DryRun bool
var Crop bool
var OutputPath string
var Rate int
var Codec string
var AudioRate int // k
var AudioCodec string
var AudioChannels int
var FileSize int // MB
var Size string
var Preset string
var Seek float64
var Duration float64
var Extension string
var DrawTitle bool

func getSafePath(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	safePath := base + ext
	_, err := os.Stat(safePath)
	for i := 1; err == nil; i++ {
		safePath = base + "." + strconv.FormatInt(int64(i), 10) + ext
		_, err = os.Stat(safePath)
	}
	return safePath
}

func NewOutputVideo(input *Video) *Video {
	output := NewVideoFromVideo(input)
	output.setSize(Size)
	output.setEncodeCodec(Codec)
	output.audioRate = 0
	output.audioChannels = 0
	output.audioCodec = "copy"
	output.rate = Rate
	output.seek = Seek
	if Duration > 0 {
		output.duration = Duration
	}
	// If audio rate is specified (and lower than input rate)
	if AudioRate > 0 && (input.audioRate == 0 || AudioRate < input.audioRate) {
		output.audioRate = AudioRate
		// default to AC3
		output.audioCodec = "ac3"
	}
	// If audio channels are specified
	if AudioChannels > 0 && AudioChannels != input.audioChannels {
		output.audioChannels = AudioChannels
		// default to AC3
		output.audioCodec = "ac3"
	}
	// If 2 audio channels default to AAC
	if output.audioCodec != "copy" && output.audioChannels == 2 {
		output.audioCodec = "aac"
	}
	// If codec is specified overrule them all
	if AudioCodec != "" {
		output.audioCodec = AudioCodec
	}
	if FileSize > 0 {
		output.setFileSize(FileSize)
	}
	if Extension != "" {
		output.extension = Extension
	}
	return output
}

func getEncodeCommand(input *Video, output *Video) *exec.Cmd {
	var args []string
	args = append(args,
		"-y", "-hide_banner",
	)
	// Start options for -i input.file
	if strings.Contains(input.codec, "cuvid") {
		args = append(args,
			"-hwaccel", "cuda", // cuda, dxva2, qsv, d3d11va, qsv, cuvid
			"-hwaccel_output_format", "cuda", // cuda, nv12, p010le, p016le
			// "-pixel_format", "yuv420p",
			// "-hwaccel", "nvdec",
			// "-hwaccel_output_format", "yuv420p",
			// "-hwaccel_output_format", "nv12",
			// "-hwaccel_output_format", "yuv420p10le",
			// "-pix_fmt", "yuv420p",
			// "-c:v", input.codec,
		)
	}
	if input.codec != "" {
		args = append(args,
			"-c:v", input.codec,
		)
	}
	if (input.cropTop + input.cropBottom + input.cropLeft + input.cropRight) > 0 {
		args = append(args,
			"-crop", (strconv.FormatInt(int64(input.cropTop), 10) +
				"x" + strconv.FormatInt(int64(input.cropBottom), 10) +
				"x" + strconv.FormatInt(int64(input.cropLeft), 10) +
				"x" + strconv.FormatInt(int64(input.cropRight), 10)),
		)
	}
	if input.width != output.width {
		args = append(args,
			"-resize", (strconv.FormatInt(int64(output.width), 10) +
				"x" + strconv.FormatInt(int64(output.height), 10)),
		)
	}
	args = append(args, "-i", input.file)
	// Start output options
	// args = append(args, "-map", "0:s?")
	// args = append(args, "-map", "0:v?")
	// args = append(args, "-map", "0:a?")
	if output.seek > 0 {
		args = append(args, "-ss", strconv.FormatFloat(output.seek, 'f', -1, 64))
	}
	if output.duration > 0 {
		args = append(args, "-t", strconv.FormatFloat(output.duration, 'f', -1, 64))
	}
	if DrawTitle {
		title := strings.ToUpper(strings.Replace(input.title, ".", " ", -1))
		drawtext := "enable='between(t,0,3)':" +
			"fontfile=/Windows/Fonts/impact.ttf:" +
			"text='" + title + "':" +
			"fontsize=72:" +
			"fontcolor=ffffff:" +
			"alpha='if(lt(t,0),0,if(lt(t,0),(t-0)/0,if(lt(t,2),1,if(lt(t,3),(1-(t-2))/1,0))))':" +
			"x=(w-text_w)/2:" +
			"y=(h-text_h)/2"
		args = append(args, "-filter_complex", ("hwdownload,format=nv12,drawtext=" + drawtext + ",hwupload_cuda"))
	}
	// Start video output options
	if output.codec != "" {
		args = append(args,
			"-c:v", output.codec,
			"-rc:v", "vbr_hq",
			"-cq:v", "20",
			"-profile:v", "main",
			"-max_muxing_queue_size", "800",
		)
	}
	if output.rate > 0 {
		args = append(args,
			"-b:v", (strconv.FormatInt(int64(output.rate), 10) + "k"),
			"-maxrate:v", (strconv.FormatInt(int64(output.rate*2), 10) + "k"),
		)
	}
	// Start audio output options
	args = append(args, "-c:a", output.audioCodec)
	if output.audioRate > 0 {
		args = append(args, "-b:a", (strconv.FormatInt(int64(output.audioRate), 10) + "k"))
	}
	if output.audioChannels > 0 {
		args = append(args, "-ac", strconv.FormatInt(int64(output.audioChannels), 10))
	}
	// Start subtitle output options
	// args = append(args, "-c:s", "copy")
	// args = append(args, "-map", "0")
	// Ouput file
	output.file = getSafePath(filepath.Join(viper.GetString("encode.OutputPath"),
		(output.baseName + "." + output.size + "." + output.extension)),
	)
	args = append(args, output.file)
	return exec.Command("ffmpeg", args...)
}

var encodeCmd = &cobra.Command{
	Use:   "encode [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Encode a video",
	Long:  "Encode a video using ffmpeg",
	Run: func(cmd *cobra.Command, args []string) {
		var ffmpegCmd *exec.Cmd
		input := NewVideoFromFile(args[0])

		if Crop {
			input.detectCrop()
		}

		if DetectVolume {
			input.detectVolume()
		}

		output := NewOutputVideo(input)

		ffmpegCmd = getEncodeCommand(input, output)

		fmt.Printf("Input file: %s\n", input.file)
		fmt.Printf("Output file: %s\n", output.file)
		fmt.Printf("File extension: %s -> %s\n", input.extension, output.extension)
		fmt.Printf("Title: %s\n", input.title)
		fmt.Printf("Year: %s\n", input.year)
		fmt.Printf("Scene info: %s\n", input.sceneInfo)
		fmt.Printf("Seek: %f\n", input.seek)
		fmt.Printf("Duration: %f\n", input.duration)
		fmt.Printf("Pixel format: %s\n", input.pixelFormat)
		fmt.Printf("Video crop top: %d\n", input.cropTop)
		fmt.Printf("Video crop bottom: %d\n", input.cropBottom)
		fmt.Printf("Video crop left: %d\n", input.cropLeft)
		fmt.Printf("Video crop right: %d\n", input.cropRight)
		fmt.Printf("Video size: %s -> %s\n", input.size, output.size)
		fmt.Printf("Video width: %d -> %d\n", input.width, output.width)
		fmt.Printf("Video height: %d -> %d\n", input.height, output.height)
		fmt.Printf("Video rate: %dk -> %dk\n", input.rate, output.rate)
		fmt.Printf("Video codec: %s -> %s\n", input.codec, output.codec)
		fmt.Printf("Audio codec: %s -> %s\n", input.audioCodec, output.audioCodec)
		fmt.Printf("Audio rate: %dk -> %dk\n", input.audioRate, output.audioRate)
		fmt.Printf("Audio channels: %d -> %d\n", input.audioChannels, output.audioChannels)
		fmt.Printf("Audio channel layout: %s\n", input.audioLayout)
		fmt.Printf("\n%+v\n\n", ffmpegCmd)

		if DryRun || DetectVolume {
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
