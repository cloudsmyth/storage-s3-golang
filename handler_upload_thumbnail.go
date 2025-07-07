package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	maxMemory := 10 << 20
	if err := r.ParseMultipartForm(int64(maxMemory)); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse file", err)
		return
	}
	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get file form", err)
		return
	}
	defer file.Close()

	contentHeader := fileHeader.Header.Get("Content-Type")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get file data", err)
		return
	}

	extension, err := cfg.getSupportedAssetType(contentHeader)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get file extension", err)
	}

	newFileName := fmt.Sprintf("%s%s", videoID, extension)

	uniquePath := filepath.Join(cfg.assetsRoot, newFileName)
	createdFile, err := os.Create(uniquePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create file", err)
		return
	}
	defer createdFile.Close()

	if _, err = io.Copy(createdFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy file data", err)
		return
	}

	videoMeta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get data from databasae", err)
		return
	}

	if videoMeta.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Incorrect user", err)
		return
	}

	dataUrl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, newFileName)
	videoMeta.ThumbnailURL = &dataUrl

	if err = cfg.db.UpdateVideo(videoMeta); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Bad database update", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMeta)
}
