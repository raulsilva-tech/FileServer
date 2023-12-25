package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
	err = requestVideo(videoDTO)
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
		Message: "None",
	})
}

func requestVideo(videoInfo dto.DownloadVideoInput) error {
	//getting the path to save the video
	cfg, _ := configs.LoadConfig(".")

	//if file exists, do not request
	if _, err := os.Stat(cfg.Directory + "/" + videoInfo.FileName); err == nil {
		fmt.Println(videoInfo.FileName + ": Already exists")
		return nil
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
			return err
		}
		defer res.Body.Close()
		d, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		//creating the video
		tmpfile, err := os.Create(cfg.Directory + "/" + videoInfo.FileName)
		if err != nil {
			return err
		}
		defer tmpfile.Close()
		//actively creating the video
		tmpfile.Write(d)

		return nil
	}
}

// func removeAllFilesFromThisUser(userId string, cfg configs.Config) error {

// 	files, err := os.ReadDir(cfg.Directory)
// 	if err != nil {
// 		return err
// 	}

// 	for _, file := range files {
// 		if strings.Contains(file.Name(), userId) {
// 			_ = os.Remove(cfg.Directory + "/" + file.Name())
// 		}
// 	}

// 	return nil
// }
