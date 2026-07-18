package main

func ffmpegInputArgs() []string {
	return []string{"-f", "avfoundation", "-i", ":0"}
}
