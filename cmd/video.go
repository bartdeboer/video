package cmd

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	"hevc": "hevc_cuvid",
	"h264": "h264_cuvid",
	// "h264": "h264_nvdec",
}

var encoders = map[string]string{
	"hevc":       "hevc_cuvid",
	"h264":       "h264_nvenc",
	"h265":       "hevc_nvenc",
	"h264_nvenc": "h264_nvenc",
	"hevc_nvenc": "hevc_nvenc",
	"libx264":    "libx264",
	"libx265":    "libx265",
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

type Video struct {
	file          string
	baseName      string
	extension     string
	width         int
	height        int
	size          string
	seek          float64
	duration      float64
	rate          int
	codec         string
	pixelFormat   string
	audioRate     int
	audioCodec    string
	audioChannels int
	audioLayout   string
	cropTop       int
	cropBottom    int
	cropLeft      int
	cropRight     int
	title         string
	year          string
	sceneInfo     string
}

func NewVideo() *Video {
	return &Video{}
}

func NewVideoFromFile(file string) *Video {
	video := NewVideo()
	video.file = file
	video.detectVideo()
	video.detectAudio()
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
	height, _ := strconv.ParseInt(strings.Trim(size, "p"), 10, 0)
	if width, ok := Sizes[int(height)]; ok {
		video.size = size
		if video.width > width {
			resizeRatio := float64(width) / float64(video.width)
			video.height = int(resizeRatio * float64(video.height))
			video.width = width
		}
	}
}

func (video *Video) setEncodeCodec(codec string) {
	video.codec = codec
	if encoder, ok := encoders[video.codec]; ok {
		video.codec = encoder
	}
}

func (video *Video) setDecodeCodec(codec string) {
	video.codec = codec
	if decoder, ok := decoders[video.codec]; ok {
		video.codec = decoder
	}
}

func (video *Video) setFileSize(fileSize int) {
	if fileSize > 0 {
		video.rate = int((float64(fileSize) * 8192 / video.duration) - float64(video.audioRate))
	}
}

func (input *Video) detectVideo() (int, int) {
	fmt.Print("Detecting video\n")
	input.extension = strings.Trim(filepath.Ext(input.file), ".")
	input.baseName = filepath.Base(strings.TrimSuffix(input.file, ("." + input.extension)))
	// r, _ := regexp.Compile("\\[^\\]]*\\]")
	// r, _ := regexp.Compile("\\.[0-9]{4}\\.(.*)$")
	r, _ := regexp.Compile("^(.*)\\.([0-9]{4})\\.(.*)$")
	submatches := r.FindStringSubmatch(input.baseName)
	if len(submatches) == 4 {
		input.title = submatches[1]
		input.year = submatches[2]
		input.sceneInfo = submatches[3]
		input.baseName = input.title + "." + input.year
	}
	input.baseName = r.ReplaceAllString(input.baseName, "")
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
	input.detectSize()
	input.duration = duration
	input.setDecodeCodec(keyValues["codec_name"])
	input.pixelFormat = keyValues["pix_fmt"]
	input.rate = int(rate / 1000)
	return int(width), int(height)
}

func (input *Video) detectAudio() {
	fmt.Print("Detecting audio\n")
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

func (input *Video) detectCrop() {
	fmt.Print("Detecting black bars\n")
	var args []string
	args = append(args,
		"-y", "-hide_banner",
	)
	if decoder, ok := decoders[input.codec]; ok {
		args = append(args,
			"-hwaccel", "cuda",
			// "-hwaccel_output_format", "cuda",
			"-c:v", decoder,
		)
	}
	args = append(args,
		"-i", input.file,
		"-vf", "fps=1/60,cropdetect=24:16:0",
		"-to", "600",
		"-an",
		"-f", "null",
		getNullDevice(),
	)
	ffmpegCmd := exec.Command("ffmpeg", args...)
	out, _ := ffmpegCmd.CombinedOutput()
	// fmt.Printf("Crop Detect: %s\n", string(out))
	r, _ := regexp.Compile("crop=([0-9]+):([0-9]+):([0-9]+):([0-9]+)")
	matches := r.FindAllStringSubmatch(string(out), -1)
	minX, minY, maxWidth, maxHeight := input.width, input.height, 0, 0
	for _, submatches := range matches {
		// fmt.Printf("%q\n", submatches[0])
		width, _ := strconv.ParseInt(submatches[1], 10, 0)
		height, _ := strconv.ParseInt(submatches[2], 10, 0)
		x, _ := strconv.ParseInt(submatches[3], 10, 0)
		y, _ := strconv.ParseInt(submatches[4], 10, 0)
		maxWidth = int(math.Max(float64(maxWidth), float64(width)))
		maxHeight = int(math.Max(float64(maxHeight), float64(height)))
		minX = int(math.Min(float64(minX), float64(x)))
		minY = int(math.Min(float64(minY), float64(y)))
	}
	input.cropTop = minY
	input.cropBottom = input.height - (minY + maxHeight)
	input.cropLeft = minX
	input.cropRight = input.width - (minX + maxWidth)
	input.height = maxHeight
	input.width = maxWidth
}

func (input *Video) detectVolume() /* float64 */ {
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
