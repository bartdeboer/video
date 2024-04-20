// Copyright 2020 Bart de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	"github.com/bartdeboer/cfg"
	"github.com/spf13/cobra"
)

var initial = Config{
	VideoStream:        0,
	AudioStream:        0,
	SubtitleStream:     0,
	ConstantQuality:    -1,
	ConstantRateFactor: -1,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "video",
	Short: "A brief description of your application",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func indexOfOsArgs(search string) int {
	for i, value := range os.Args {
		if value == (search) {
			return i
		}
	}
	return -1
}

func init() {

	cfg.BindPersistentFlagsKey("encode", rootCmd, &initial)

	fmt.Printf("%v\n", initial)

	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(bulkCmd)

	if argIndex := indexOfOsArgs("--preset"); argIndex != -1 {
		valueIndex := argIndex + 1
		if valueIndex < len(os.Args) {
			initial.Preset = os.Args[valueIndex]
		}
	}

	switch initial.Preset {
	case "":
		break
	case "telegram-small":
		initial.Codec = "libx264"
		// initial.Size = "1080p"
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.DrawTitle = false
		initial.Extension = "mp4"
		// initial.ConstantQuality = 23 // 1080p:19 720p:23
		initial.ConstantRateFactor = 26
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = false
		// initial.WatermarkFile = "watermark-small.png"
		// initial.WatermarkPosition = "W-w-48:48"
		break
	case "telegram-fair":
		initial.Codec = "libx264"
		initial.Size = "1080p"
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.DrawTitle = false
		initial.Extension = "mp4"
		// initial.ConstantQuality = 23 // 1080p:19 720p:23
		initial.ConstantRateFactor = 23
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = false
		// initial.WatermarkFile = "watermark-small.png"
		// initial.WatermarkPosition = "W-w-48:48"
		break
	case "telegram":
		initial.Codec = "h264_nvenc"
		initial.Size = "1080p"
		initial.FileSize = 2016 // max 2048
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.DrawTitle = true
		initial.Extension = "mp4"
		initial.ConstantQuality = 19 // 1080p:19 720p:23
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = true
		break
	case "telegram-hevc":
		initial.Codec = "hevc_nvenc"
		initial.Size = "1080p"
		initial.FileSize = 2016 // max 2048
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.DrawTitle = true
		initial.Extension = "mp4"
		initial.ConstantQuality = 22 // 1080p:19 720p:23
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = true
	case "telegram-x265":
		initial.Codec = "libx265"
		initial.Size = "1080p"
		initial.FileSize = 2016 // max 2048
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.DrawTitle = true
		initial.Extension = "mp4"
		// initial.ConstantRateFactor = 26
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = true
		break
	case "phone":
		// Size = "720p"
		// FileSize = 1490 // max 1536
		initial.AudioRate = 196
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		// DrawTitle = true
		initial.Extension = "mp4"
	case "homevideo":
		initial.Codec = "libx265"
		// initial.AudioRate = 196
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.ConstantRateFactor = 21
		// DrawTitle = true
		initial.Extension = "mp4"
	case "teams":
		initial.Codec = "h264_nvenc"
		initial.Size = "1080p"
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.AudioStream = 0
		initial.Extension = "mp4"
		initial.ConstantQuality = 27 // 1080p:19 720p:23
		initial.PixelFormat = "yuv420p"
		initial.ColorTransfer = "bt709"
		initial.OptMetadata = true
		break
	default:
		fmt.Print("Unknown preset\n")
		os.Exit(0)
	}

}
