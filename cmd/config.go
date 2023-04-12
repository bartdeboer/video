package cmd

type Config struct {
	Preset             string  `usage:"Preset (telegram, phone)"`
	DetectVolume       bool    `usage:"Detect volume"`
	Volume             string  `usage:"Set volume level"`
	DryRun             bool    `usage:"Dry run"`
	Crop               bool    `usage:"Autocrop black bars"`
	OutputPath         string  `usage:"Output path"`
	Rate               int     `usage:"(ffmpeg b:v) Video bitrate (k)"`
	Codec              string  `usage:"(ffmpeg c:v) Video codec"`
	InputCodec         string  `usage:"Input decoder codec"`
	VideoStream        int     `usage:"Audio stream index to use"`
	AudioRate          int     `usage:"(ffmpeg b:a) Audio bitrate (k)"`
	AudioCodec         string  `usage:"(ffmpeg c:a) Audio codec"`
	AudioChannels      int     `usage:"Number of audio channels"`
	AudioStream        int     `usage:"Audio stream index to use"`
	AudioDelay         float64 `usage:"Audio stream delay (seconds)"`
	FileSize           int     `usage:"Target file size (MB)"`
	Size               string  `usage:"Resolution (480p, 576p, 720p, 1080p, 1440p or 2160p)"`
	Seek               float64 `usage:"Seek (seconds)"`
	Duration           float64 `usage:"Duration (seconds)"`
	Ss                 string  `usage:"Seek (hh:mm:ss.xxx)"`
	To                 string  `usage:"To (hh:mm:ss.xxx)"`
	Extension          string  `usage:"File extension"`
	DrawTitle          bool    `usage:"Draw title (requires reencode)"`
	Title              string  `usage:"Video title to draw"`
	FontFile           string  `usage:"Font file"`
	BurnSubtitles      bool    `usage:"Hardcodes the subtitles"`
	BurnImageSubtitles bool    `usage:"Hardcodes the subtitle images"`
	SubtitleStream     int     `usage:"Subtitle stream index to use"`
	ConstantQuality    int     `usage:"Constant Quality (0-63)"`
	ConstantRateFactor int     `usage:"Constant Rate Factor (0-51)"`
	FfmpegPath         string  `usage:"Path containing the ffmpeg binary"`
	PixelFormat        string  `usage:"Pixel format (yuv420p, yuv420p10le, ...)"`
	ColorTransfer      string  `usage:"Color transfer (smpte2084, bt709, ...)"`
	Denoise            bool    `usage:"Removes film grain"`
	OptMetadata        bool    `usage:"Optimize metadata"`
	TwoPass            bool    `usage:"Perform 2-pass encoding"`
	Tonemap            string  `usage:"tonemap (mobius, hable, ...)"`
	Tune               string  `usage:"tune (animation, film, ...)"`
	Level              string  `usage:"level (3, 4.1, ...)"`
}
