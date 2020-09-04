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

type Config struct {
	Preset          string  `usage:"Preset (telegram, phone)"`
	DetectVolume    bool    `usage:"Detect volume"`
	Volume          string  `usage:"Set volume level"`
	DryRun          bool    `usage:"Dry run"`
	Crop            bool    `usage:"Autocrop black bars"`
	OutputPath      string  `usage:"Output path"`
	Rate            int     `usage:"(ffmpeg b:v) Video bitrate (k)"`
	Codec           string  `usage:"(ffmpeg c:v) Video codec"`
	VideoStream     int     `usage:"Audio stream index to use"`
	AudioRate       int     `usage:"(ffmpeg b:a) Audio bitrate (k)"`
	AudioCodec      string  `usage:"(ffmpeg c:a) Audio codec"`
	AudioChannels   int     `usage:"Number of audio channels"`
	AudioStream     int     `usage:"Audio stream index to use"`
	FileSize        int     `usage:"Target file size (MB)"`
	Size            string  `usage:"Resolution (480p, 576p, 720p, 1080p, 1440p or 2160p)"`
	Seek            float64 `usage:"Seek (seconds)"`
	Duration        float64 `usage:"Duration (seconds)"`
	Extension       string  `usage:"File extension"`
	DrawTitle       bool    `usage:"Draw title (requires reencode)"`
	Title           string  `usage:"Video title to draw"`
	FontFile        string  `usage:"Font file"`
	BurnSubtitles   bool    `usage:"Hardcodes the subtitles"`
	SubtitleStream  int     `usage:"Subtitle stream index to use"`
	ConstantQuality int     `usage:"Constant Quality (0-63)"`
}

var initial = Config{
	VideoStream:    -1,
	AudioStream:    -1,
	SubtitleStream: 0,
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

	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(bulkCmd)

	if argIndex := indexOfOsArgs("--preset"); argIndex != -1 {
		valueIndex := argIndex + 1
		if valueIndex < len(os.Args) {
			initial.Preset = os.Args[valueIndex]
		}
	}

	if initial.Preset == "telegram" {
		initial.Size = "720p"
		// initial.FileSize = 1490 // max 1536
		initial.FileSize = 2016 // max 2048
		initial.AudioRate = 144 // 128 = good
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.DrawTitle = true
		initial.Extension = "mp4"
	}

	if initial.Preset == "phone" {
		// Size = "720p"
		// FileSize = 1490 // max 1536
		initial.AudioRate = 196
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		// DrawTitle = true
		initial.Extension = "mp4"
	}
}
