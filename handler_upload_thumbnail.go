package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	const maxMemory = 10 << 20 // 10 MB
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	//save to disk
	mediaTypeSplit := strings.Split(mediaType,`/`)
	file_type := ""
	if len(mediaTypeSplit) == 2 {
		file_type = mediaTypeSplit[1]
	}
	rndStr := make([]byte, 32)
	_, err = rand.Read(rndStr)
	if err != nil {
		return
	}
	base64fileName := base64.RawURLEncoding.EncodeToString(rndStr)
	file_name := base64fileName + "." + file_type
	videoPath := filepath.Join(cfg.assetsRoot,file_name)
	videoUrl := "http://localhost:8091/assets/" + file_name
	


	diskFile, err := os.Create(videoPath)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to save video", err)
		return
	}
	defer diskFile.Close()

	_, err = io.Copy(diskFile,file)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to save video", err)
		return
	}
	video.ThumbnailURL = &videoUrl

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
