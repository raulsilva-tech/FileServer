package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/raulsilva-tech/FileServer/configs"
	"github.com/raulsilva-tech/FileServer/internal/dto"
)

type VideoHandler struct{}

type Error struct {
	Message string `json:"message"`
}

func NewVideoHandler() *VideoHandler {
	return &VideoHandler{}
}

func (vh *VideoHandler) EraseVideos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userId := chi.URLParam(r, "id")
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Message: "id is required",
		})
		return
	}

	//getting the path to save the video
	cfg, _ := configs.LoadConfig(".")

	err := removeAllFilesFromThisUser(userId, *cfg)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Error{
		Message: "Success",
	})

}

func (vh *VideoHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	//getting what has been sent in the body request
	var videoDTO dto.DownloadVideoInput
	err := json.NewDecoder(r.Body).Decode(&videoDTO)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Message: err.Error(),
		})
		return
	}

	//sending request to get the video from the DVR
	_, err = requestAndSaveVideo(videoDTO)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Error{
		Message: "Success",
	})
}

func requestAndSaveVideo(videoInfo dto.DownloadVideoInput) (*os.File, error) {
	//getting the path to save the video
	cfg, _ := configs.LoadConfig(".")

	//if file exists, do not request
	if _, err := os.Stat(cfg.Directory + "/" + videoInfo.FileName); err == nil {
		fmt.Println(videoInfo.FileName + ": Already exists")
		file, _ := os.Open(cfg.Directory + "/" + videoInfo.FileName)
		return file, err
	} else {

		req, err := http.NewRequest(http.MethodGet, videoInfo.Url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		client := &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		d, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		//creating the video
		tmpfile, err := os.Create(cfg.Directory + "/" + videoInfo.FileName)
		if err != nil {
			return nil, err
		}
		defer tmpfile.Close()
		//actively creating the video
		tmpfile.Write(d)

		return tmpfile, err
	}
}

func removeAllFilesFromThisUser(userId string, cfg configs.Config) error {

	files, err := os.ReadDir(cfg.Directory)
	if err != nil {
		return err
	}

	files_found := false

	for _, file := range files {
		if strings.Contains(file.Name(), "U_"+userId) {
			files_found = true
			_ = os.Remove(cfg.Directory + "/" + file.Name())
		}
	}

	if !files_found {
		return errors.New("No files found with user id " + userId)
	}

	return nil
}
