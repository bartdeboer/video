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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "video",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
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
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.video.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	initConfig()

	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(bulkCmd)

	if argIndex := indexOfOsArgs("--preset"); argIndex != -1 {
		valueIndex := argIndex + 1
		if valueIndex < len(os.Args) {
			Preset = os.Args[valueIndex]
		}
	}

	Codec = viper.GetString("encode.Codec")
	OutputPath = viper.GetString("encode.OutputPath")
	Extension = viper.GetString("encode.Extension")
	FontFile = viper.GetString("encode.FontFile")

	if Preset == "telegram" {
		Size = "720p"
		FileSize = 1490 // max 1536
		AudioRate = 128
		AudioChannels = 2
		AudioCodec = "aac"
		DrawTitle = true
		Extension = "mp4"
	}

	if Preset == "phone" {
		// Size = "720p"
		// FileSize = 1490 // max 1536
		AudioRate = 196
		AudioChannels = 2
		AudioCodec = "aac"
		// DrawTitle = true
		Extension = "mp4"
	}

	rootCmd.PersistentFlags().StringVarP(&Preset, "preset", "p", Preset, "Preset (telegram)")
	encodeCmd.Flags().BoolVarP(&DetectVolume, "detect-volume", "", DetectVolume, "Detect volume")
	encodeCmd.Flags().BoolVarP(&DryRun, "dry-run", "", DryRun, "Dry Run")
	encodeCmd.Flags().BoolVarP(&Crop, "crop", "c", Crop, "Crop black bars")
	encodeCmd.Flags().IntVarP(&FileSize, "file-size", "f", FileSize, "Target file size (MB)")
	encodeCmd.Flags().StringVarP(&Codec, "codec:v", "", Codec, "(ffmpeg c:v) Video codec")
	encodeCmd.Flags().IntVarP(&Rate, "bitrate:v", "", Rate, "(ffmpeg b:v) Video bitrate (k)")
	encodeCmd.Flags().StringVarP(&AudioCodec, "codec:a", "", AudioCodec, "(ffmpeg c:a) Audio codec")
	encodeCmd.Flags().IntVarP(&AudioRate, "bitrate:a", "", AudioRate, "(ffmpeg b:a) Audio bitrate (k)")
	encodeCmd.Flags().IntVarP(&AudioChannels, "audio-channels", "", AudioChannels, "Number of output audio channels")
	encodeCmd.Flags().StringVarP(&Size, "size", "s", Size, "Resolution (480p, 576p, 720p, 1080p, 1440p or 2160p)")
	encodeCmd.Flags().Float64VarP(&Seek, "seek", "", Seek, "Seek (seconds)")
	encodeCmd.Flags().Float64VarP(&Duration, "duration", "", Duration, "Duration (seconds)")
	encodeCmd.Flags().StringVarP(&Extension, "extension", "", Extension, "File extension")
	encodeCmd.Flags().BoolVarP(&DrawTitle, "draw-title", "", DrawTitle, "Draw Title")

	// rootCmd.Flags().StringVarP(&configFile, "configFile", "c", "", fmt.Sprintf("config file (default is ~/%s.%s)", defaultConfigFilename, defaultConfigExt))

	// encodeCmd.Flags().StringVarP(&OutputPath, "output-path", "", "", fmt.Sprintf("Output path (default is %s)", viper.GetString("encode.OutputPath")))
	rootCmd.PersistentFlags().StringVarP(&OutputPath, "output-path", "", OutputPath, "Output path")

	// viper.BindPFlag("encode.OutputPath", encodeCmd.Flags().Lookup("output-path"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".video" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".video")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// viper.SetConfigType("yaml")
		// viper.Set("encode.OutputPath", "")
		// viper.Set("encode.Codec", "h264_nvenc")
		// viper.Set("encode.Extension", "mp4")
		// viper.Set("encode.FontFile", "/Windows/Fonts/impact.ttf")
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
		// fmt.Println("Write default config")
		// if err := viper.SafeWriteConfig(); err != nil {
		// 	fmt.Println(err)
		// }
	}
}
