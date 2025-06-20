package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	//db
	db_video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to Find Video.", err)
		return
	}
	if db_video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not the owner of video file. Unauthorized", nil)
		return
	}


	const maxMemory = 1 << 30 // 1000 MB
	
	r.Body = http.MaxBytesReader(w,r.Body,maxMemory)

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusRequestEntityTooLarge, "File too large or invalid multipart form", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	fileType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}
	if fileType != "video/mp4"{
		respondWithError(w, http.StatusBadRequest, "Invalid Media type (expecting mp4)", nil)
		return
	}

	tmpFile, err := os.CreateTemp(cfg.filepathRoot,"tubely-upload.mp4")
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to save video", err)
		return
	}
	tmpFile.Seek(0, io.SeekStart)

	processedPath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to process the video file.", err)
		return
	}
	defer os.Remove(processedPath)
	processedFile, err := os.Open(processedPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to open processed video file", err)
		return
	}
	defer processedFile.Close()

	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to determin aspect ratio", err)
		return
	}
	var prefix string
	switch aspectRatio{
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	default:
		prefix = "other"
	}
	rndStr := make([]byte, 32)
	_, err = rand.Read(rndStr)
	if err != nil {
		return
	}

	//file processing
	


	file_name := prefix + "/" + base64.RawURLEncoding.EncodeToString(rndStr) + ".mp4"
	_, err = cfg.s3Client.PutObject(r.Context(),&s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &file_name,
		Body: processedFile,
		ContentType: &fileType,
	})
	if err != nil {
		fmt.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}
	// fmt.Printf("S3 upload succeeded: %+v\n", putOutput)
	// s3Url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",cfg.s3Bucket, cfg.s3Region, file_name)
	// s3Url := cfg.s3Bucket + "," + file_name
	s3Url := cfg.s3CfDistribution + file_name
	db_video.VideoURL = &s3Url

	err = cfg.db.UpdateVideo(db_video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}
	// fmt.Println("Normal Exit")
	// video, err := cfg.dbVideoToSignedVideo(db_video)
	// if err != nil {
	// 	fmt.Println(err)
	// 	respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
	// 	return
	// }
	respondWithJSON(w, http.StatusOK, db_video)
}


