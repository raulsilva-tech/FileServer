package main

import (
	"log"
	"net/http"

	"github.com/raulsilva-tech/FileServer/configs"
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

	//creating a route for the file server
	http.Handle("/", http.FileServer(http.Dir(cfg.Directory)))
	log.Printf("Serving %s on HTTP port: %s\n", cfg.Directory, cfg.Port)

	//starting the server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
