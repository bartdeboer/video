package main

import (
	"fmt"
	"os"

	"github.com/bartdeboer/flag"
)

var initial = Config{
	VideoStream:        0,
	AudioStream:        0,
	SubtitleStream:     0,
	ConstantQuality:    -1,
	ConstantRateFactor: -1,
}

func SetPreset(preset string) {
	switch preset {
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
	case "homevideo2":
		initial.Codec = "hevc_nvenc"
		// initial.AudioRate = 196
		initial.AudioChannels = 2
		initial.AudioCodec = "aac"
		initial.ConstantQuality = 22
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

func SetInitial() ([]string, error) {

	if err := flag.SetDefaults(&initial); err != nil {
		return nil, fmt.Errorf("error setting default values: %v", err)
	}

	if err := flag.ParseEnv(&initial); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %v", err)
	}

	yamlCfg := struct {
		Encode *Config `yaml:"encode"`
	}{
		Encode: &initial,
	}

	if err := LoadYaml(&yamlCfg); err != nil {
		return nil, err
	}

	args, flags := flag.ParseArgs(os.Args[1:])

	{
		_, helpExists := flags["help"]
		_, hExists := flags["h"]
		if helpExists || hExists {
			flag.PrintDefaults(&initial)
			os.Exit(0)
		}
	}

	if preset, exists := flags["preset"]; exists {
		SetPreset(preset)
	}

	err := flag.SetFlags(&initial, flags)
	if err != nil {
		return nil, fmt.Errorf("error parsing command-line arguments: %v", err)
	}

	return args, nil
}

func main() {

	args, err := SetInitial()
	if err != nil {
		fmt.Printf("error initializing: %v", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		flag.PrintDefaults(&initial)
		os.Exit(1)
	}

	switch args[0] {
	case "encode":
		encode(args[1])
	case "bulk":

	}

}
