package cmd

import (
	"github.com/spf13/cobra"
)

var vesCmd = &cobra.Command{
	Use:   "ves [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Encode a VES file",
	Long:  "Encode a VES file using ffmpeg",
	Run: func(cmd *cobra.Command, args []string) {

		// var ffmpegCmd *exec.Cmd
		// input := NewVideoFromFile(args[0])

		// input.detectVideo(initial.VideoStream)
		// input.detectAudio(initial.AudioStream)

		// if initial.Crop {
		// 	input.detectCrop()
		// }

		// if initial.DetectVolume {
		// 	input.detectVolume()
		// }

		// if initial.InputCodec != "" {
		// 	input.codec = initial.InputCodec
		// }

		// output := input.NewOutputVideoFromCmdAgrs()
		// ffmpegCmd = input.getEncodeCommand(output)

		// if initial.DryRun {
		// 	os.Exit(0)
		// }

		// ffmpegCmd.Stdout = os.Stdout
		// ffmpegCmd.Stderr = os.Stderr

		// err := ffmpegCmd.Run()
		// if err != nil {
		// 	log.Fatalf("ffmpegCmd.Run() failed with %s\n", err)
		// }
	},
}
