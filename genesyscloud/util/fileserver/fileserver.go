package fileserver

import (
	"context"
	"log"
	"net/http"
	"sync"
)

func StartHttpServer(waitGroup *sync.WaitGroup, directory, port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}

	http.DefaultServeMux = new(http.ServeMux)
	http.Handle("/", http.FileServer(http.Dir(directory)))

	go func() {
		defer waitGroup.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe(): %v", err)
		}
		log.Println("Finished serving")
	}()

	return srv
}

func ShutDown(server *http.Server, waitGroup *sync.WaitGroup) {
	if err := server.Shutdown(context.TODO()); err != nil {
		log.Println("Error shutting down server:", err)
	}
	waitGroup.Wait()
}
