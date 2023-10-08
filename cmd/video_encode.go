// mpv:
//  * pixel format: yuv420p10 (same as pix_fmt in ffprobe, yuv420p10le)
//  * primaries: bt.2020 (same as color_primaries in ffprobe, bt2020)
//  * colormatrix: bt.2020-ncl (same as color_space in ffprobe, bt2020nc)
//  * levels: limited (same as color_range in ffprobe, tv)
//  * gamma: pq (same as color_transfer in ffprobe, smpte2084)

// ffprobe:
// * pix_fmt: yuv420p10le
// * color_range: tv
// * color_space: bt2020nc
// * color_transfer: smpte2084
// * color_primaries: bt2020

// As you can see, the information is consistent across both tools, but they
// use different terminology. mpv doesn't specifically list the color space
// because it lists the color matrix, which is a part of the color space.
// Conversely, ffprobe doesn't list the color matrix because it lists the color
// space, which includes the color matrix information.

// Your current code checks the color_transfer value (smpte2084) to detect if
// it's an HDR video and needs to be tonemapped into SDR, which is a correct
// approach. The color transfer function (gamma) is one of the key components
// that differentiates HDR and SDR content.

package cmd

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func (input *Video) getEncodeCommand(output *Video) *exec.Cmd {
	var args []string
	filters := []string{}
	swFilters := []string{}
	openClFilters := []string{}
	isHwAcceleratedDecode := false

	args = append(args,
		"-y", "-hide_banner",
	)

	// Start input stream options:

	// GPU decoding:
	if strings.Contains(input.codec, "cuvid") || strings.Contains(input.codec, "nvenc") {
		// http://ffmpeg.org/pipermail/ffmpeg-devel/2018-November/235929.html
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

		isHwAcceleratedDecode = true

		// prepend hwdownload for decode filters
		swFilters = append([]string{
			fmt.Sprintf("hwdownload"),
		}, swFilters...)
	}

	// Input stream decoder:
	if input.codec != "" && input.codec != "ffmpeg" && input.codec != "copy" {
		args = append(args,
			"-c:v", input.codec,
		)
	}

	// Input stream crop options
	if isHwAcceleratedDecode && (input.cropTop+input.cropBottom+input.cropLeft+input.cropRight) > 0 {
		args = append(args,
			"-crop", fmt.Sprintf("%dx%dx%dx%d", input.cropTop, input.cropBottom, input.cropLeft, input.cropTop),
		)
	}

	if !isHwAcceleratedDecode && (input.cropTop+input.cropBottom+input.cropLeft+input.cropRight) > 0 {
		swFilters = append(swFilters,
			fmt.Sprintf("crop=%d:%d:%d:%d", input.width, input.height, input.cropLeft, input.cropTop),
		)
	}

	// Input stream resize options
	if isHwAcceleratedDecode && (input.width != output.width || input.height != output.height) {
		args = append(args,
			"-resize", (strconv.FormatInt(int64(output.width), 10) +
				"x" + strconv.FormatInt(int64(output.height), 10)),
		)
	}

	if !isHwAcceleratedDecode && (input.width != output.width || input.height != output.height) {
		swFilters = append(swFilters, ("scale=" +
			strconv.FormatInt(int64(output.width), 10) + ":" +
			strconv.FormatInt(int64(output.height), 10)),
		)
	}

	// fmt.Printf("HEIGHT: %d -> %d\n", output.height, output.height%16)
	// fmt.Printf("WIDTH: %d -> %d\n", output.width, output.width%16)

	// fmt.Print("HEIGHT", output.height, "\n")

	_, hasStandardHeight := Sizes[output.height]

	// Input stream pad options
	if !hasStandardHeight && ((output.height%16) > 0 || (output.width%16) > 0) {
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

	if initial.WatermarkFile != "" {
		args = append(args, "-i", initial.WatermarkFile)
	}

	if output.audioDelay != 0 {
		args = append(args, "-itsoffset", strconv.FormatFloat(output.audioDelay, 'f', -1, 64), "-i", input.file)
		output.audioInput = 1
	}

	if initial.OptMetadata {
		args = append(args, "-map_metadata", "-1")
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

	if initial.WatermarkFile != "" {
		if initial.WatermarkPosition != "" {
			// top-right: W-w-48:48
			// bottom-left: 48:H-h-48
			filters = append(filters, fmt.Sprintf("[v][1:v:0]overlay=%s[v]", initial.WatermarkPosition))
		} else {
			filters = append(filters, "[v][1:v:0]overlay[v]", initial.WatermarkPosition)
		}
	}

	videoStream := "0:v"
	if initial.VideoStream > -1 {
		videoStream = fmt.Sprintf("0:v:%d", initial.VideoStream)
	}

	switch input.pixelFormat {
	case "yuv420p10le", "yuv422p10le", "yuv444p10le":
		swFilters = append(swFilters,
			"format=p010le",
		)
	default:
		swFilters = append(swFilters,
			// "format=yuv420p",
			"format=nv12",
		)
	}

	if initial.Denoise {
		// zscale=transfer=bt709,
		// format=nv12,
		// format=yuv420p,
		// openClFilters = append(openClFilters,
		// 	"nlmeans_opencl=s=1:p=7:pc=5:r=5:rc=5",
		// )
		swFilters = append(swFilters,
			"nlmeans=s=1:p=7:pc=5:r=5:rc=5",
		)
	}

	if input.colorTransfer != "bt709" && output.colorTransfer == "bt709" {
		switch input.colorTransfer {
		case "bt709":
			swFilters = append(swFilters,
				"format=yuv420p",
			)
		case "arib-std-b67":
			fallthrough
		case "smpte2084":
			tonemap := output.tonemap
			if tonemap == "" {
				tonemap = "mobius"
			}
			// if tonemap == "drm" {
			// 	swFilters = append(swFilters,
			// 		"tonemap=mantiuk:contrast=1.5:desat=0.0",
			// 	)
			// } else {
			openClFilters = append(openClFilters,
				// Software tonal map:
				// "zscale=t=linear", "format=gbrpf32le", "zscale=p=bt709", "tonemap=tonemap=hable", "zscale=t=bt709:m=bt709:r=tv",
				// Hardware tonal map:
				// tonemap=tm=drm:type=mantiuk:contrast=1.5:desat=0.0:format=nv12
				fmt.Sprintf("tonemap_opencl=tonemap=%s:param=0.01:desat=0.0:range=tv:primaries=bt709:transfer=bt709:matrix=bt709:format=nv12", tonemap),
				// "tonemap_opencl=tonemap=mobius:param=0.01:desat=0.0:range=tv:primaries=bt709:transfer=bt709:matrix=bt709:format=nv12",
			)
			// }
			output.colorPrimaries = "bt709"
			output.colorSpace = "bt709"
		case "bt601":
			fallthrough
		case "unknown":
			swFilters = append(swFilters,
				// "colorspace=space=bt709:trc=bt709", // :primaries=bt709
				// "scale=in_color_matrix=bt601:out_color_matrix=bt709",
				// "colorspace=all=bt709",
				// "zscale=tin=bt601:pin=bt601:min=bt601:transfer=bt709:matrix=bt709:primaries=bt709",
				"colormatrix=bt601:bt709",
			)
		}
	}

	// Covert HLG HDR (arib-std-b67)
	if input.colorTransfer != "smpte2084" && output.colorTransfer == "smpte2084" {
		if output.codec == "libx265" {
			swFilters = append(swFilters,
				"zscale=transfer=smpte2084",
			)
			args = append(args,
				"-pix_fmt", "yuv420p10le",
				"-x265-params", "hdr-opt=1:repeat-headers=1:colorprim=bt2020:transfer=smpte2084:colormatrix=bt2020nc",
			)
		}
		if output.codec == "hevc_nvenc" {
			swFilters = append(swFilters,
				"zscale=transfer=smpte2084,format=p010le",
			)
		}
	}

	if len(openClFilters) > 0 {
		// args = append(args, "-init_hw_device", "opencl=ocl", "-filter_hw_device", "ocl")
		args = append(args, "-init_hw_device", "opencl=gpu:0.0", "-filter_hw_device", "gpu")

		// filters = append([]string{
		// 	fmt.Sprintf("[v]format=yuv420p,hwupload,%s,hwdownload,format=yuv420p[v]", strings.Join(openClFilters, ",")),
		// }, filters...)

		if isHwAcceleratedDecode && len(swFilters) == 2 && swFilters[0] == "hwdownload" && swFilters[1] == "format=nv12" {
			swFilters = append([]string{},
				fmt.Sprintf("%s,hwdownload,format=nv12", strings.Join(openClFilters, ",")),
			)

			// 	swFilters = append([]string{
			// 		"hwdownload",
			// 		swFilters[1],
			// 		"hwupload",
			// 	},
			// 		fmt.Sprintf("%s,hwdownload,format=nv12", strings.Join(openClFilters, ",")),
			// 	)
		} else {
			swFilters = append(swFilters,
				fmt.Sprintf("hwupload,%s,hwdownload,format=nv12", strings.Join(openClFilters, ",")),
			)
		}

	}

	if output.codec == "copy" {
		// args = append([]string{
		// 	"-fflags", "+igndts",
		// }, args...)
		args = append(args, "-map", videoStream)
	} else {
		if len(swFilters) > 0 {
			// append nv12 (fixed)
			// swFilters = append(swFilters,
			// 	"format=nv12",
			// )

			filters = append([]string{
				fmt.Sprintf("[%s]%s[v]", videoStream, strings.Join(swFilters, ",")),
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
	}

	// Start video output options
	if output.codec == "h264_nvenc" {
		// ffmpeg -y -vsync 0 -hwaccel cuda -hwaccel_output_format cuda -i input.mp4 -c:a copy
		// -c:v h264_nvenc -preset p6 -tune hq -b:v 5M -bufsize 5M -maxrate 10M -qmin 0 -g 250
		// -bf 3 -b_ref_mode middle -temporal-aq 1 -rc-lookahead 20 -i_qfactor 0.75 -b_qfactor
		// 1.1 output.mp4
		args = append(args,
			"-c:v", output.codec,
			"-preset:v", "p7", // p1 ... p7, fast, medium, slow
			// "-profile:v", "main",
			// "-level:v", "4.1", // auto, 1 ... 6.2
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
			"-preset:v", "p7", // p1 ... p7, fast, medium, slow
			"-level:v", "4.1", // auto, 1 ... 6.2
			"-rc:v", "vbr", // vbr, vbr_hq, cbr
			"-rc-lookahead:v", "32",
			"-bf:v", "4", // 3
			"-bufsize:v", "16M", // 8M
			"-max_muxing_queue_size", "800",
		)
	} else if output.codec == "libx265" || output.codec == "libx264" {
		args = append(args,
			"-c:v", output.codec,
			"-preset:v", "slow",
		)
	} else if output.codec != "" {
		args = append(args, "-c:v", output.codec)
	}
	if output.constantRateFactor != -1 {
		args = append(args, "-crf:v", strconv.FormatInt(int64(output.constantRateFactor), 10))
	} else if output.constantQuality != -1 {
		args = append(args, "-cq:v", strconv.FormatInt(int64(output.constantQuality), 10))
	}
	if output.rate > 0 {
		args = append(args,
			// "-minrate:v", (strconv.FormatInt(int64(math.Round(float64(output.rate)*float64(0.5))), 10) + "k"),
			"-b:v", (strconv.FormatInt(int64(output.rate), 10) + "k"),
		)
		if !initial.TwoPass {
			args = append(args,
				"-maxrate:v", (strconv.FormatInt(int64(math.Round(float64(output.rate)*float64(1.0))), 10) + "k"),
			)
		}
	}

	// Start audio output options
	if input.audioStream != -1 {
		audioStream := fmt.Sprintf("%d:a", output.audioInput)
		if initial.AudioStream > -1 {
			audioStream = fmt.Sprintf("%d:a:%d", output.audioInput, initial.AudioStream)
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
	}

	if initial.Tune != "" {
		args = append(args, "-tune", initial.Tune)
	}

	if initial.Level != "" {
		args = append(args, "-level:v", initial.Level)
	}

	// Start subtitle output options
	// args = append(args, "-c:s", "copy")
	// args = append(args, "-map", "0")
	// Ouput file
	output.file = getSafePath(filepath.Join(initial.OutputPath,
		(output.baseName + "." + output.size + "." + output.extension)),
	)

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

	if !initial.DryRun && initial.TwoPass && output.codec == "libx265" {
		var pass1Args = append(args,
			"-x265-params", "no-slow-firstpass=1:pass=1",
			"-an",
			"-f", "null",
			getNullDevice(),
		)
		pass1ffmpegCmd := exec.Command(cmdName, pass1Args...)
		fmt.Printf("\n%+v\n\n", pass1ffmpegCmd)
		pass1ffmpegCmd.Stdout = os.Stdout
		pass1ffmpegCmd.Stderr = os.Stderr
		err := pass1ffmpegCmd.Run()
		if err != nil {
			log.Fatalf("pass1ffmpegCmd.Run() failed with %s\n", err)
		}

		args = append(args,
			"-x265-params", "pass=2",
		)
	}

	args = append(args, output.file)

	ffmpegCmd := exec.Command(cmdName, args...)

	fmt.Printf("\n%+v\n\n", ffmpegCmd)

	return ffmpegCmd
}
