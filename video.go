package main

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var Sizes = map[int]int{
	480:  720,
	576:  720,
	720:  1280,
	1080: 1920,
	1440: 2560,
	2160: 3840,
}

var decoders = map[string]string{
	"hevc":       "hevc_cuvid", // sw: "hevc"
	"h264":       "h264_cuvid",
	"h263":       "h263_cuvid",
	"mpeg4":      "mpeg4_cuvid",
	"mpeg2":      "mpeg2_cuvid",
	"mpeg2video": "mpeg2_cuvid",
	"mpeg1":      "mpeg1_cuvid",
	"vc1":        "vc1_cuvid",
	"vp9":        "vp9_cuvid",
	// "h264": "h264_nvdec",
}

var encoders = map[string]string{
	"hevc":       "hevc_nvenc",
	"h264":       "h264_nvenc",
	"h265":       "hevc_nvenc",
	"h264_nvenc": "h264_nvenc",
	"hevc_nvenc": "hevc_nvenc",
	"libx264":    "libx264",
	"libx265":    "libx265",
}

type Video struct {
	file               string
	baseName           string
	extension          string
	width              int
	height             int
	size               string
	seek               float64
	duration           float64
	stream             int
	rate               int
	codec              string
	pixelFormat        string
	colorRange         string
	colorSpace         string
	colorTransfer      string
	colorPrimaries     string
	audioStream        int
	audioRate          int
	audioCodec         string
	audioChannels      int
	audioLayout        string
	audioDelay         float64
	audioInput         int
	cropTop            int
	cropBottom         int
	cropLeft           int
	cropRight          int
	title              string
	year               string
	extraInfo          string
	volume             string
	constantQuality    int
	constantRateFactor int
	tonemap            string
}

func NewVideo() *Video {
	return &Video{
		constantQuality:    -1,
		constantRateFactor: -1,
	}
}

func NewVideoFromFile(file string) *Video {
	video := NewVideo()
	video.file = file
	return video
}

func NewVideoFromVideo(source *Video) *Video {
	video := *source
	return &video
}

func (video *Video) detectSize() {
	for height, width := range Sizes {
		if width == video.width || height == video.height {
			video.size = strconv.FormatInt(int64(height), 10) + "p"
		}
	}
}

func (video *Video) setSize(size string) {
	height, _ := strconv.Atoi(strings.Trim(size, "p"))
	if width, ok := Sizes[height]; ok {
		video.size = size
		aspectRatio := float64(video.width) / float64(video.height)

		if video.width > width {
			video.width = width
			video.height = int(math.RoundToEven(float64(video.width) / aspectRatio))
		}

		if video.height > height {
			video.height = height
			video.width = int(math.RoundToEven(float64(video.height) * aspectRatio))
		}
	}
}

func (video *Video) setEncodeCodec(codec string) {
	video.codec = codec
	if encoder, ok := encoders[video.codec]; ok {
		video.codec = encoder
	}
}

func (video *Video) getDecoder(codec string) string {
	if initial.Decoder != "" {
		return initial.Decoder
	}
	if decoder, ok := decoders[video.codec]; ok {
		return decoder
	}
	return codec
}

func (video *Video) setFileSize(fileSize int) {
	if fileSize > 0 {
		video.rate = int((float64(fileSize) * 8192 / video.duration) - float64(video.audioRate))
	}
}

func (input *Video) detectVideo(streamIndex int) (int, int) {
	fmt.Print("Detecting video...\n")
	input.stream = streamIndex
	input.extension = strings.Trim(filepath.Ext(input.file), ".")
	input.baseName = filepath.Base(strings.TrimSuffix(input.file, ("." + input.extension)))
	// r, _ := regexp.Compile("\\[^\\]]*\\]")
	// r, _ := regexp.Compile("\\.[0-9]{4}\\.(.*)$")

	// Auto simplify titles
	// r, _ := regexp.Compile("^(.*)[. ]([0-9]{4})[. ](.*)$")
	// submatches := r.FindStringSubmatch(input.baseName)
	// if len(submatches) == 4 {
	// 	input.title = submatches[1]
	// 	input.year = submatches[2]
	// 	input.extraInfo = submatches[3]
	// 	input.baseName = input.title + "." + input.year
	// }

	// I don't know what the purpose was of this:
	// input.baseName = r.ReplaceAllString(input.baseName, "")

	var cmdName = "ffprobe"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffprobe.exe")
	}
	ffprobCmd := exec.Command(cmdName,
		"-v", "error",
		"-select_streams", "v:"+strconv.Itoa(streamIndex),
		"-show_format",
		"-show_streams",
		// "-show_entries", "stream=width,height",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd, "=")
	if err != nil {
		log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
	}
	width, _ := strconv.ParseInt(keyValues["width"], 10, 0)
	height, _ := strconv.ParseInt(keyValues["height"], 10, 0)
	duration, _ := strconv.ParseFloat(keyValues["duration"], 64)
	rate, _ := strconv.ParseInt(keyValues["bit_rate"], 10, 0)
	input.width = int(width)
	input.height = int(height)
	input.detectSize()
	input.duration = duration
	input.codec = input.getDecoder(keyValues["codec_name"])
	input.pixelFormat = keyValues["pix_fmt"]
	input.colorRange = keyValues["color_range"]
	input.colorSpace = keyValues["color_space"]
	input.colorTransfer = keyValues["color_transfer"]
	input.colorPrimaries = keyValues["color_primaries"]
	input.rate = int(rate / 1000)
	return int(width), int(height)
}

func (input *Video) detectAudio(streamIndex int) {
	fmt.Print("Detecting audio...\n")
	input.audioStream = streamIndex
	var cmdName = "ffprobe"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffprobe.exe")
	}
	ffprobCmd := exec.Command(cmdName,
		"-v", "error",
		"-select_streams", "a:"+strconv.Itoa(streamIndex),
		"-show_streams",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd, "=")
	if err != nil {
		log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
	}
	if len(keyValues) == 0 {
		input.audioStream = -1
		return
	}
	rate, _ := strconv.ParseInt(keyValues["bit_rate"], 10, 0)
	channels, _ := strconv.ParseInt(keyValues["channels"], 10, 0)
	input.audioCodec = keyValues["codec_name"]
	input.audioRate = int(rate / 1000)
	input.audioChannels = int(channels)
	input.audioLayout = keyValues["channel_layout"]
}

func (input *Video) detectCrop() {
	fmt.Print("Detecting black bars...\n")
	var args []string

	detectDuration := 600.0

	if initial.CropDetectDuration != 0 {
		detectDuration = initial.CropDetectDuration
	} else {
		if initial.Duration != 0 && initial.Duration < detectDuration {
			detectDuration = initial.Duration
		}
		if initial.To != "" {
			to, _ := parseTimeStringToSeconds(initial.To)
			detectDuration = math.Min(to, detectDuration) - 2
		}
	}

	// detectDuration := math.Min(math.Min(initial.Duration, input.duration), 600)
	// frameInterval := int(math.Round(detectDuration / 9))
	// fpsValue := float64(10) / detectDuration
	// filter := fmt.Sprintf("fps=fps=%.6f,cropdetect=0.1:16:0", fpsValue)
	filter := fmt.Sprintf("fps=fps=10/%.6f,cropdetect=0.1:16:0", detectDuration)
	// filter := fmt.Sprintf("select='not(lt(mod(t,%d),1))',cropdetect=0.1:16:0", frameInterval)

	args = append(args,
		"-y", "-hide_banner",
	)
	if strings.Contains(input.codec, "cuvid") || strings.Contains(input.codec, "nvenc") {
		args = append(args,
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
		)
	}
	args = append(args,
		"-c:v", input.codec,
	)
	args = append(args,
		"-i", input.file,
		"-vf", filter,
		"-to", strconv.Itoa(int(detectDuration)),
		"-an",
		"-f", "null",
		getNullDevice(),
	)
	var cmdName = "ffmpeg"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffmpeg.exe")
	}

	ffmpegCmd := exec.Command(cmdName, args...)

	fmt.Printf("\n%+v\n\n", ffmpegCmd)

	out, _ := ffmpegCmd.CombinedOutput()
	fmt.Printf("Crop Detect: %s\n", string(out))
	r, _ := regexp.Compile("crop=([0-9]+):([0-9]+):([0-9]+):([0-9]+)")

	matches := r.FindAllStringSubmatch(string(out), -1)
	minX, minY, maxRight, maxBottom, maxWidth, maxHeight := input.width, input.height, 0, 0, 0, 0
	for _, submatches := range matches {
		x, _ := strconv.ParseInt(submatches[3], 10, 0)
		y, _ := strconv.ParseInt(submatches[4], 10, 0)
		width, _ := strconv.ParseInt(submatches[1], 10, 0)
		height, _ := strconv.ParseInt(submatches[2], 10, 0)
		maxWidth = int(math.Max(float64(maxWidth), float64(width)))
		maxHeight = int(math.Max(float64(maxHeight), float64(height)))
		minX = int(math.Min(float64(minX), float64(x)))
		minY = int(math.Min(float64(minY), float64(y)))
		maxRight = int(math.Max(float64(maxRight), float64(x+width)))
		maxBottom = int(math.Max(float64(maxBottom), float64(y+height)))
	}
	input.cropTop = minY
	input.cropBottom = input.height - (minY + maxHeight)
	// input.cropBottom = input.height - maxBottom
	input.cropLeft = minX
	input.cropRight = input.width - (minX + maxWidth)
	// input.cropRight = input.width - maxRight
	input.height = maxHeight
	input.width = maxWidth

	// os.Exit(0)

}

func (input *Video) detectVolume() /* float64 */ {
	fmt.Print("Detecting volume levels...\n")
	var cmdName = "ffmpeg"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffmpeg.exe")
	}
	ffmpegCmd := exec.Command(cmdName,
		"-hide_banner",
		"-i", input.file,
		// "-to", "400",
		"-vn",
		"-filter:a", "volumedetect",
		"-f", "null",
		getNullDevice(),
	)
	// keyValues, err := getKeyValuesFromCommand(ffmpegCmd, ":")
	// if err != nil {
	// 	log.Fatalf("getKeyValuesFromCommand() failed with %s\n", err)
	// }
	// input.volume = keyValues["max_volume"]
	out, _ := ffmpegCmd.CombinedOutput()
	// fmt.Println(string(out))
	r, _ := regexp.Compile("max_volume:[^\\n]+")
	_, value := getKeyStringValue(r.FindString(string(out)), ":")
	input.volume = value
	// fmt.Println(r.FindString(string(out)))
}

func (input *Video) NewOutputVideoFromCmdAgrs() *Video {
	output := NewVideoFromVideo(input)
	output.setSize(initial.Size)
	output.setEncodeCodec(initial.Codec)
	if output.codec == "copy" {
		input.codec = "copy"
		initial.ConstantRateFactor = -1
		initial.ConstantQuality = -1
		initial.Duration = 0
		initial.Rate = -1
		initial.FileSize = -1
		initial.Duration = -1
	}

	// fmt.Print("INPUT CODEC::::::::", input.codec, "\n")
	// fmt.Print("INITIAL CODEC::::::::", initial.Codec, "\n")
	// fmt.Print("OUTPUT CODEC::::::::", output.codec, "\n")

	output.audioCodec = "copy"
	output.rate = initial.Rate
	output.seek = initial.Seek
	if initial.Ss != "" {
		output.seek, _ = parseTimeStringToSeconds(initial.Ss)
	}
	if initial.PixelFormat != "" {
		output.pixelFormat = initial.PixelFormat
	}
	if initial.ColorTransfer != "" {
		output.colorTransfer = initial.ColorTransfer
	}
	if initial.ConstantRateFactor != -1 {
		output.constantRateFactor = initial.ConstantRateFactor
	} else if initial.ConstantQuality != -1 {
		output.constantQuality = initial.ConstantQuality
	}
	if initial.Duration > 0 {
		output.duration = initial.Duration
	} else if initial.To != "" {
		to, _ := parseTimeStringToSeconds(initial.To)
		output.duration = to - output.seek
	}
	if output.duration+output.seek > input.duration {
		output.duration = input.duration - output.seek
	}
	// If audio rate is specified (only override if less than input rate)
	if initial.AudioRate > 0 && (input.audioRate == 0 || initial.AudioRate <= input.audioRate) {
		output.audioRate = initial.AudioRate
		// default to AC3
		output.audioCodec = "ac3"
	}
	// If audio channels are specified (only override if less than input channels)
	if initial.AudioChannels > 0 && initial.AudioChannels <= input.audioChannels {
		output.audioChannels = initial.AudioChannels
		// default to AC3
		output.audioCodec = "ac3"
	}
	// If we're not copying and if we have 2 audio channels: Default to AAC
	if output.audioCodec != "copy" && output.audioChannels == 2 {
		output.audioCodec = "aac"
	}
	// If codec is specified overrule them all
	if initial.AudioCodec != "" {
		output.audioCodec = initial.AudioCodec
	}
	// If output audio is the same as input audio just copy the stream
	if input.audioRate == output.audioRate &&
		input.audioChannels == output.audioChannels &&
		input.audioCodec == output.audioCodec &&
		!initial.DetectVolume {
		output.audioCodec = "copy"
	}
	if initial.DetectVolume && input.volume != "" {
		output.volume = strings.Trim(output.volume, "-")
	}
	if initial.FileSize > 0 {
		output.setFileSize(initial.FileSize)
	}
	if initial.Extension != "" {
		output.extension = initial.Extension
	}
	output.tonemap = initial.Tonemap
	// if initial.ConstantQuality > 0 {
	// 	output.constantQuality = initial.ConstantQuality
	// }
	output.audioDelay = initial.AudioDelay
	return output
}
