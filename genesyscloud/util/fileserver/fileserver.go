package fileserver

import (
	"log"
	"net/http"
	"sync"
)

func StartHttpServer(wg *sync.WaitGroup, directory, port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}

	http.DefaultServeMux = new(http.ServeMux)
	http.Handle("/", http.FileServer(http.Dir(directory)))

	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe(): %v", err)
		}
		log.Println("Finished serving")
	}()

	return srv
}
