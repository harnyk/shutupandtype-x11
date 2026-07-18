package main

func ffmpegInputArgs() []string {
	return []string{"-f", "alsa", "-i", "default"}
}
