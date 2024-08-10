package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func getNullDevice() string {
	if runtime.GOOS == "windows" {
		return "NUL"
	}
	return "/dev/null"
}

func parseTimeStringToSeconds(timeStr string) (float64, error) {
	parts := strings.Split(timeStr, ":")
	var durationStr string

	switch len(parts) {
	case 1:
		durationStr = parts[0] + "s"
	case 2:
		durationStr = parts[0] + "m" + parts[1] + "s"
	case 3:
		durationStr = parts[0] + "h" + parts[1] + "m" + parts[2] + "s"
	default:
		return 0, fmt.Errorf("invalid time format")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, err
	}

	return duration.Seconds(), nil
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

	fmt.Printf("\n%+v\n\n", cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("cmd.StdoutPipe() failed with %s", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cmd.Start() failed with %s", err)
	}

	defer cmd.Wait() // Ensures that the command finishes and resources are cleaned up

	scanner := bufio.NewScanner(stdout)
	keyValues := make(map[string]string)

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		key, value := getKeyStringValue(text, sep)
		keyValues[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %s", err)
	}

	return keyValues, nil
}

// func getKeyValuesFromCommand(cmd *exec.Cmd, sep string) (map[string]string, error) {
// 	stdout, err := cmd.StdoutPipe()
// 	// stdout, err := cmd.StderrPipe()
// 	if err != nil {
// 		log.Fatalf("cmd.StdoutPipe() failed with %s\n", err)
// 	}
// 	keyValues := map[string]string{}
// 	scanner := bufio.NewScanner(stdout)

// 	fmt.Printf("\n%+v\n\n", cmd)

// 	cmd.Start()
// 	for scanner.Scan() {
// 		text := strings.TrimSpace(scanner.Text())
// 		if text == "" {
// 			continue
// 		}
// 		fmt.Printf("%s\n", text)
// 		key, value := getKeyStringValue(text, sep)
// 		keyValues[key] = value
// 	}
// 	return keyValues, scanner.Err()
// }

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
