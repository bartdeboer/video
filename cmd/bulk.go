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
			output := NewOutputVideo(input)

			if input.codec == "h264_cuvid" {
				output.codec = "copy"
			}

			ffmpegCmd := getEncodeCommand(input, output)

			fmt.Printf("Input file: %s\n", input.file)
			fmt.Printf("Output file: %s\n", output.file)
			fmt.Printf("File extension: %s -> %s\n", input.extension, output.extension)
			fmt.Printf("Title: %s\n", input.title)
			fmt.Printf("Year: %s\n", input.year)
			fmt.Printf("Scene info: %s\n", input.sceneInfo)
			fmt.Printf("Seek: %f\n", input.seek)
			fmt.Printf("Duration: %f\n", input.duration)
			fmt.Printf("Pixel format: %s\n", input.pixelFormat)
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

			fmt.Println(filePath)

			if DryRun || DetectVolume {
				continue
			}

			ffmpegCmd.Stdout = os.Stdout
			ffmpegCmd.Stderr = os.Stderr

			err := ffmpegCmd.Run()
			if err != nil {
				log.Fatalf("ffmpegCmd.Run() failed with %s\n", err)
			}
		}

		// input := NewVideoFromFile(args[0])

	},
}
