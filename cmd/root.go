package cmd

import (
	"fmt"
	"os"

	"github.com/bartdeboer/mystack/mw/cfg"
	"github.com/spf13/cobra"
)

type Config struct {
	Preset        string  `usage:"Preset (telegram, phone)"`
	DetectVolume  bool    `usage:"Detect volume"`
	DryRun        bool    `usage:"Dry run"`
	Crop          bool    `usage:"Autocrop black bars"`
	OutputPath    string  `usage:"Output path"`
	Rate          int     `usage:"(ffmpeg b:v) Video bitrate (k)"`
	Codec         string  `usage:"(ffmpeg c:v) Video codec"`
	AudioRate     int     `usage:"(ffmpeg b:a) Audio bitrate (k)"`
	AudioCodec    string  `usage:"(ffmpeg c:a) Audio codec"`
	AudioChannels int     `usage:"Number of audio channels"`
	FileSize      int     `usage:"Target file size (MB)"`
	Size          string  `usage:"Resolution (480p, 576p, 720p, 1080p, 1440p or 2160p)"`
	Seek          float64 `usage:"Seek (seconds)"`
	Duration      float64 `usage:"Duration (seconds)"`
	Extension     string  `usage:"File extension"`
	DrawTitle     bool    `usage:"Draw title (requires reencode)"`
	FontFile      string  `usage:"Font file"`
}

var initial Config

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

	cfg.BindPersistentFlags(rootCmd, "encode", &initial)

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
		initial.FileSize = 1490 // max 1536
		initial.AudioRate = 128
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
