package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func reduceRatio(width, height int) (int, int) {
	divisor := gcd(width, height)
	return width / divisor, height / divisor
}

func getVideoAspectRatio(filepath string) (string, error) {
	// fmt.Println(filepath)
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	var b, e bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &e
	// fmt.Println("Starting command")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error Running command:", cmd.String())
		fmt.Println("Stderr:", e.String())
		return "", err
	}
	var output struct {
		Streams []struct {
			Width 			int 		`json:"width"`
			Height 			int 		`json:"height"`
			DisplayRatio 	string 		`json:"display_aspect_ratio"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(b.Bytes(), &output); err != nil {
		return "", fmt.Errorf("error parsing ffprome output: %s", err)
	}
	// fmt.Println(reduceRatio(output.Streams[0].Width, output.Streams[0].Height))
	// width,height := reduceRatio(output.Streams[0].Width, output.Streams[0].Height)
	// ratio := fmt.Sprintf("%d:%d",width,height)
	return output.Streams[0].DisplayRatio, nil
}


func processVideoForFastStart(filepath string) (string, error) {
	outputFilePath := filepath + ".processing"
	// fmt.Println(filepath)
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	var b, e bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &e
	// fmt.Println("Starting command")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error Running command:", cmd.String())
		fmt.Println("Stderr:", e.String())
		return "", err
	}



	return outputFilePath, nil
}



// func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error){
// 	presignedClient := s3.NewPresignClient(s3Client)
// 	obj, err := presignedClient.PresignGetObject(context.Background(),&s3.GetObjectInput{Bucket: &bucket,Key: &key},s3.WithPresignExpires(expireTime))
// 	if err != nil {
// 		return "", err
// 	}
// 	return obj.URL, nil
// }

// func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
// 	// orignal_url := video.VideoURL
// 	if video.VideoURL == nil {
// 		return video, nil
// 	}
// 	// fmt.Println(*video.VideoURL)
// 	url := strings.Split(*video.VideoURL, ",")
// 	// fmt.Println(url)
// 	if len(url) <2 {
// 		return video, nil

// 	}
// 	bucket := url[0]
// 	key := url[1]
// 	signed_url, err := generatePresignedURL(cfg.s3Client,bucket,key,time.Hour*2)
// 	if err != nil {
// 		return video, err
// 	}
// 	video.VideoURL = &signed_url
// 	// fmt.Println(signed_url)
// 	return video, nil

// }