// Copyright 2020 Bart de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
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
	"hevc":  "hevc_cuvid", // sw: "hevc"
	"h264":  "h264_cuvid",
	"h263":  "h263_cuvid",
	"mpeg4": "mpeg4_cuvid",
	"mpeg2": "mpeg2_cuvid",
	"mpeg1": "mpeg1_cuvid",
	"vc1":   "vc1_cuvid",
	"vp9":   "vp9_cuvid",
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

func getKeyStringValue(input string, sep string) (string, string) {
	arr := strings.SplitN(string(input), sep, 2)
	if len(arr) == 2 {
		return strings.TrimSpace(arr[0]), strings.TrimSpace(arr[1])
	}
	return strings.TrimSpace(arr[0]), ""
}

func getKeyIntValue(input string, sep string) (string, int, error) {
	arr := strings.SplitN(string(input), sep, 2)
	key := arr[0]
	value, err := strconv.ParseInt(arr[1], 10, 0)
	return key, int(value), err
}

func getKeyValuesFromCommand(cmd *exec.Cmd, sep string) (map[string]string, error) {
	stdout, err := cmd.StdoutPipe()
	// stdout, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("cmd.StdoutPipe() failed with %s\n", err)
	}
	keyValues := map[string]string{}
	scanner := bufio.NewScanner(stdout)
	cmd.Start()
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		// fmt.Printf("%s\n", text)
		key, value := getKeyStringValue(text, sep)
		keyValues[key] = value
	}
	return keyValues, scanner.Err()
}

type Video struct {
	file            string
	baseName        string
	extension       string
	width           int
	height          int
	size            string
	seek            float64
	duration        float64
	rate            int
	codec           string
	pixelFormat     string
	colorRange      string
	colorSpace      string
	colorTransfer   string
	colorPrimaries  string
	audioRate       int
	audioCodec      string
	audioChannels   int
	audioLayout     string
	cropTop         int
	cropBottom      int
	cropLeft        int
	cropRight       int
	title           string
	year            string
	extraInfo       string
	volume          string
	constantQuality int
}

func NewVideo() *Video {
	return &Video{
		constantQuality: -1,
	}
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
			video.height = int(math.RoundToEven(resizeRatio*float64(video.height)/2) * 2)
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
		input.extraInfo = submatches[3]
		input.baseName = input.title + "." + input.year
	}

	// I don't know what the purpose was of this:
	// input.baseName = r.ReplaceAllString(input.baseName, "")

	var cmdName = "ffprobe"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffprobe.exe")
	}
	ffprobCmd := exec.Command(cmdName,
		"-v", "error",
		"-select_streams", "v:0",
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
	input.setDecodeCodec(keyValues["codec_name"])
	input.pixelFormat = keyValues["pix_fmt"]
	input.colorRange = keyValues["color_range"]
	input.colorSpace = keyValues["color_space"]
	input.colorTransfer = keyValues["color_transfer"]
	input.colorPrimaries = keyValues["color_primaries"]
	input.rate = int(rate / 1000)
	return int(width), int(height)
}

func (input *Video) detectAudio() {
	fmt.Print("Detecting audio\n")
	var cmdName = "ffprobe"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffprobe.exe")
	}
	ffprobCmd := exec.Command(cmdName,
		"-v", "error",
		"-select_streams", "a:0",
		"-show_streams",
		"-of", "default=noprint_wrappers=1",
		"-i", input.file,
	)
	keyValues, err := getKeyValuesFromCommand(ffprobCmd, "=")
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
			"-hwaccel_output_format", "cuda",
			"-c:v", decoder,
		)
	}
	args = append(args,
		"-i", input.file,
		"-vf", "fps=1/60,cropdetect=0.1:16:0",
		"-to", "600",
		"-an",
		"-f", "null",
		getNullDevice(),
	)
	var cmdName = "ffmpeg"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffmpeg.exe")
	}
	ffmpegCmd := exec.Command(cmdName, args...)
	out, _ := ffmpegCmd.CombinedOutput()
	fmt.Printf("Crop Detect: %s\n", string(out))
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

func (input *Video) NewOutputVideo() *Video {
	output := NewVideoFromVideo(input)
	output.setSize(initial.Size)
	output.setEncodeCodec(initial.Codec)
	output.audioCodec = "copy"
	output.rate = initial.Rate
	output.seek = initial.Seek
	if initial.PixelFormat != "" {
		output.pixelFormat = initial.PixelFormat
	}
	if initial.ColorTransfer != "" {
		output.colorTransfer = initial.ColorTransfer
	}
	if initial.ConstantQuality != -1 {
		output.constantQuality = initial.ConstantQuality
	}
	if initial.Duration > 0 {
		output.duration = initial.Duration
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
	if initial.ConstantQuality > 0 {
		output.constantQuality = initial.ConstantQuality
	}
	return output
}

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

func (input *Video) getEncodeCommand(output *Video) *exec.Cmd {
	var args []string
	filters := []string{}
	decFilters := []string{}
	openClFilters := []string{}

	args = append(args,
		"-y", "-hide_banner",
	)

	// Start input stream options:

	// GPU decoding:
	if strings.Contains(input.codec, "cuvid") {
		args = append(args,
			"-hwaccel", "cuda", // nvdec, cuda, dxva2, qsv, d3d11va, qsv, cuvid
			"-hwaccel_output_format", "cuda", // cuda, nv12, p010le, p016le
			// "-pixel_format", "yuv420p",
			// "-hwaccel", "nvdec",
			// "-hwaccel_output_format", "yuv420p",
			// "-hwaccel_output_format", "nv12",
			// "-hwaccel_output_format", "yuv420p10le",
			// "-pix_fmt", "yuv420p",
			// "-c:v", input.codec,
		)

		// prepend hwdownload for decode filters
		decFilters = append([]string{
			fmt.Sprintf("hwdownload"),
		}, decFilters...)
	}
	// Input stream decoder:
	if input.codec != "" {
		args = append(args,
			"-c:v", input.codec,
		)
	}
	// Input stream crop options
	if (input.cropTop + input.cropBottom + input.cropLeft + input.cropRight) > 0 {
		args = append(args,
			"-crop", (strconv.FormatInt(int64(input.cropTop), 10) +
				"x" + strconv.FormatInt(int64(input.cropBottom), 10) +
				"x" + strconv.FormatInt(int64(input.cropLeft), 10) +
				"x" + strconv.FormatInt(int64(input.cropRight), 10)),
		)
	}
	// Input stream resize options
	if input.width != output.width {
		args = append(args,
			"-resize", (strconv.FormatInt(int64(output.width), 10) +
				"x" + strconv.FormatInt(int64(output.height), 10)),
		)
	}

	// fmt.Printf("HEIGHT: %d -> %d\n", output.height, output.height%16)
	// fmt.Printf("WIDTH: %d -> %d\n", output.width, output.width%16)

	_, hasStandardHeight := Sizes[output.height]

	// Input stream pad options
	if !hasStandardHeight && (output.height%16) > 0 || (output.width%16) > 0 {
		padWidth := int(math.Ceil((float64(output.width) / float64(16))) * 16)
		padHeight := int(math.Ceil((float64(output.height) / float64(16))) * 16)
		filters = append(filters, fmt.Sprintf("[v]pad=%d:%d:%d:%d,setsar=1[v]",
			padWidth,
			padHeight,
			int((padWidth-output.width)/2),
			int((padHeight-output.height)/2),
		))
		output.width = padWidth
		output.height = padHeight
	}
	// Input stream file location
	if input.file != "" {
		args = append(args, "-i", input.file)
	}

	// Start output stream options:

	// args = append(args, "-map", "0:s?")
	// args = append(args, "-map", "0:v?")
	// args = append(args, "-map", "0:a?")
	if output.seek > 0 {
		args = append(args, "-ss", strconv.FormatFloat(output.seek, 'f', -1, 64))
	}
	if output.duration > 0 {
		args = append(args, "-t", strconv.FormatFloat(output.duration, 'f', -1, 64))
	}

	if initial.DrawTitle {
		title := strings.ToUpper(strings.Replace(output.title, ".", " ", -1))
		if initial.Title != "" {
			title = initial.Title
		}
		textFadeInStart := int(output.seek)
		textFadeIn := 0
		textDisplay := 2
		textFadeOut := 1
		textDisplayStart := textFadeInStart + textFadeIn
		textFadeOutStart := textFadeInStart + textFadeIn + textDisplay
		textEnd := textFadeInStart + textFadeIn + textDisplay + textFadeOut
		filters = append(filters, fmt.Sprintf("[v]drawtext=enable='between(t,%[1]d,%[6]d)'"+
			":fontfile=%[7]s"+
			":text='%[8]s'"+
			":fontsize=(w/17)"+ // 72
			":fontcolor=ffffff"+
			// ":alpha='if(lt(t,0),0,if(lt(t,0),(t-0)/0,if(lt(t,2),1,if(lt(t,3),(1-(t-2))/1,0))))'"+
			":alpha='if(lt(t,%[1]d),0,if(lt(t,%[4]d),(t-%[1]d)/%[2]d,if(lt(t,%[5]d),1,if(lt(t,%[6]d),(%[3]d-(t-%[5]d))/%[3]d,0))))'"+
			":x=(w-text_w)/2"+
			":y=(h-text_h)/2"+
			"[v]", textFadeInStart, textFadeIn, textFadeOut, textDisplayStart, textFadeOutStart, textEnd, initial.FontFile, title))
	}

	if initial.BurnSubtitles {
		subFile := input.file
		srtFile := strings.TrimSuffix(input.file, ("."+input.extension)) + ".srt"
		if _, err := os.Stat(srtFile); err == nil {
			subFile = srtFile
		}
		subFile = strings.ReplaceAll(subFile, "\\", "/")
		subFile = strings.ReplaceAll(subFile, ":/", "\\:/")
		filters = append(filters, fmt.Sprintf("[v]subtitles='%s'"+
			":stream_index=%d"+
			":force_style='Fontname=Arial,Shadow=0,Fontsize=16'"+
			"[v]", subFile, initial.SubtitleStream))
	} else if initial.BurnImageSubtitles {
		// filters = append(filters, fmt.Sprintf("[0:s:%d]scale=%d:-1[s]", initial.SubtitleStream, output.width))
		filters = append(filters, fmt.Sprintf("[0:s:%d]scale=%d:%d[s]", initial.SubtitleStream, output.width, output.height))
		filters = append(filters, "[v][s]overlay[v]")
	}

	videoStream := "0:v"
	if initial.VideoStream > -1 {
		videoStream = fmt.Sprintf("0:v:%d", initial.VideoStream)
	}

	switch input.pixelFormat {
	case "yuv420p10le", "yuv422p10le", "yuv444p10le":
		if output.pixelFormat == "yuv420p" {
			decFilters = append(decFilters,
				"format=p010le",
			)
		}
	}

	if input.colorTransfer != "unknown" && output.colorTransfer == "bt709" {
		switch input.colorTransfer {
		case "smpte2084":
			openClFilters = append(openClFilters,
				// Software tonal map:
				// "zscale=t=linear", "format=gbrpf32le", "zscale=p=bt709", "tonemap=tonemap=hable", "zscale=t=bt709:m=bt709:r=tv",
				// Hardware tonal map:
				"tonemap_opencl=tonemap=mobius:param=0.01:desat=0.0:range=tv:primaries=bt709:transfer=bt709:matrix=bt709:format=nv12",
			)
			output.colorPrimaries = "bt709"
			output.colorSpace = "bt709"
		}
	}

	if initial.Denoise {
		openClFilters = append(openClFilters,
			"nlmeans_opencl=s=1:p=7:pc=5:r=5:rc=5",
		)
	}

	if len(openClFilters) > 0 {
		// args = append(args, "-init_hw_device", "opencl=ocl", "-filter_hw_device", "ocl")
		args = append(args, "-init_hw_device", "opencl=gpu:0.0", "-filter_hw_device", "gpu")

		filters = append([]string{
			fmt.Sprintf("[v]format=yuv420p,hwupload,%s,hwdownload,format=yuv420p[v]", strings.Join(openClFilters, ",")),
		}, filters...)
	}

	if len(decFilters) > 0 {
		// append nv12 (fixed)
		decFilters = append(decFilters,
			"format=nv12",
		)

		filters = append([]string{
			fmt.Sprintf("[%s]%s[v]", videoStream, strings.Join(decFilters, ",")),
		}, filters...)
	}

	// This might be weird
	if len(filters) > 0 {
		// GPU encoding:
		if strings.Contains(output.codec, "nvenc") {
			filters = append(filters, "[v]hwupload_cuda[v]")
		}
		args = append(args, "-filter_complex", strings.Join(filters, ","))
		args = append(args, "-map", "[v]")
	} else {
		args = append(args, "-map", videoStream)
	}

	// Start video output options
	if output.codec == "h264_nvenc" {
		args = append(args,
			"-c:v", output.codec,
			"-preset:v", "p7", // p1 ... p7, fast, medium, slow
			// "-profile:v", "main",
			"-level:v", "4.1", // auto, 1 ... 6.2
			"-rc:v", "vbr", // vbr, vbr_hq, cbr
			"-bf:v", "4", // 3
			// "-refs:v", "16",
			"-b_ref_mode:v", "middle",
			"-rc-lookahead:v", "32",
			"-bufsize:v", "16M", // 8M
			"-max_muxing_queue_size", "800",
		)
	} else if output.codec == "hevc_nvenc" {
		args = append(args,
			"-c:v", output.codec,
			"-preset:v", "slow", // p1 ... p7, fast, medium, slow
			"-level:v", "4.1", // auto, 1 ... 6.2
			"-rc:v", "vbr", // vbr, vbr_hq, cbr
			"-rc-lookahead:v", "32",
			"-bufsize:v", "16M", // 8M
			"-max_muxing_queue_size", "800",
		)
	} else if output.codec != "" {
		args = append(args, "-c:v", output.codec)
	}
	if output.constantQuality != -1 {
		args = append(args,
			// "-cq:v", "24", // lower is better
			"-cq:v", strconv.FormatInt(int64(output.constantQuality), 10),
		)
	}
	if output.rate > 0 {
		args = append(args,
			"-minrate:v", (strconv.FormatInt(int64(math.Round(float64(output.rate)*float64(0.5))), 10) + "k"),
			"-b:v", (strconv.FormatInt(int64(output.rate), 10) + "k"),
			"-maxrate:v", (strconv.FormatInt(int64(math.Round(float64(output.rate)*float64(1.0))), 10) + "k"),
		)
	}

	// Start audio output options
	audioStream := "0:a"
	if initial.AudioStream > -1 {
		audioStream = fmt.Sprintf("0:a:%d", initial.AudioStream)
	}
	args = append(args, "-map", audioStream)
	args = append(args, "-c:a", output.audioCodec)
	if output.audioCodec != "copy" {
		if output.audioRate > 0 {
			args = append(args, "-b:a", (strconv.FormatInt(int64(output.audioRate), 10) + "k"))
		}
		if output.audioChannels > 0 {
			args = append(args, "-ac", strconv.FormatInt(int64(output.audioChannels), 10))
		}
		if output.volume != "" {
			args = append(args, "-filter:a", fmt.Sprintf("volume=%s", strings.Replace(output.volume, " ", "", -1)))
		}
	}
	// Start subtitle output options
	// args = append(args, "-c:s", "copy")
	// args = append(args, "-map", "0")
	// Ouput file
	output.file = getSafePath(filepath.Join(initial.OutputPath,
		(output.baseName + "." + output.size + "." + output.extension)),
	)
	args = append(args, output.file)

	fmt.Printf("Input file: %s\n", input.file)
	fmt.Printf("Output file: %s\n", output.file)
	fmt.Printf("File extension: %s -> %s\n", input.extension, output.extension)
	fmt.Printf("Title: %s\n", input.title)
	fmt.Printf("Year: %s\n", input.year)
	fmt.Printf("Extra info: %s\n", input.extraInfo)
	fmt.Printf("Seek: %f -> %f\n", input.seek, output.seek)
	fmt.Printf("Duration: %f -> %f\n", input.duration, output.duration)
	fmt.Printf("Pixel format: %s -> %s\n", input.pixelFormat, output.pixelFormat)
	fmt.Printf("Color range: %s -> %s\n", input.colorRange, output.colorRange)
	fmt.Printf("Pixel space: %s -> %s\n", input.colorSpace, output.colorSpace)
	fmt.Printf("Color transfer: %s -> %s\n", input.colorTransfer, output.colorTransfer)
	fmt.Printf("Color primaries: %s -> %s\n", input.colorPrimaries, output.colorPrimaries)
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
	fmt.Printf("Audio volume: %s -> %s\n", input.volume, output.volume)

	var cmdName = "ffmpeg"
	if initial.FfmpegPath != "" {
		cmdName = filepath.Join(initial.FfmpegPath, "ffmpeg.exe")
	}
	ffmpegCmd := exec.Command(cmdName, args...)

	fmt.Printf("\n%+v\n\n", ffmpegCmd)

	return ffmpegCmd
}
