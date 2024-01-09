package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/raulsilva-tech/FileServer/configs"
	"github.com/raulsilva-tech/FileServer/internal/webserver/handlers"
)

func main() {

	//getting parameters that set port and directory
	// port := flag.String("p", "8888", "Porta para servir os arquivos")
	// directory := flag.String("d", "./videos", "O diretório que deve ser servido")
	// flag.Parse()

	//loading configuration
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//creating a route for the file server
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "videos"))
	FileServer(r, "/", filesDir)

	vh := handlers.NewVideoHandler()

	//creating route for the DVR request to get the video
	r.Post("/get_video", vh.DownloadVideo)

	r.Get("/erase_videos/{id}", vh.EraseVideos)

	log.Printf("Serving %s on HTTP port: %s \n", cfg.Directory, cfg.Port)
	//starting the server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))

}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
