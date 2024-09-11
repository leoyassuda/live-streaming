package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Live stream app running!")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/dash", serveContent)
	http.HandleFunc("/hls", serveContent)
	http.HandleFunc("/generate", generateStreamingContent)
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveContent(writer http.ResponseWriter, request *http.Request) {
	filepath := filepath.Join(".", request.URL.Path)
	http.ServeFile(writer, request, filepath)
}

func generateStreamingContent(writer http.ResponseWriter, request *http.Request) {
	videoFile := request.URL.Query().Get("video")
	if videoFile == "" {
		http.Error(writer, "Video file not specified", http.StatusBadRequest)
		return
	}

	inputPath := filepath.Join("video", videoFile)
	dashOutputPath := filepath.Join("dash", filepath.Base(videoFile))
	hlsOutputPath := filepath.Join("hls", filepath.Base(videoFile))

	dashCmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-profile:v", "main",
		"-vf", "scale=-2:720",
		"-c:v", "libx264",
		"-b:v", "2M",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "dash",
		"-use_timeline", "1",
		"-use_template", "1",
		"-min_seg_duration", "2000000",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-init_seg_name", fmt.Sprintf("init_$RepresentationID$.m4s"),
		"-media_seg_name", fmt.Sprintf("chunk_$RepresentationID$_$Number%%05d$.m4s"),
		filepath.Join(dashOutputPath, "manifest.mpd"))

	err := dashCmd.Run()
	if err != nil {
		log.Printf("Error generating DASH content: %v", err)
		http.Error(writer, "Error generating DASH content", http.StatusInternalServerError)
	}

	hlsCmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-profile:v", "main",
		"-vf", "scale=-2:720",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(hlsOutputPath, "segment%03d.ts"),
		filepath.Join(hlsOutputPath, "playlist.m3u8"))

	err = hlsCmd.Run()
	if err != nil {
		log.Printf("Error generating HLS content: %v", err)
		http.Error(writer, "Error generating HLS content", http.StatusInternalServerError)
	}

	fmt.Fprintf(writer, "Streaming content generated for %s", videoFile)
}
