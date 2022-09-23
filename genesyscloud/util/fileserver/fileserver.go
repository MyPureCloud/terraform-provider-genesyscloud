package fileserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func Start(waitGroup *sync.WaitGroup, directory string, port int) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	http.DefaultServeMux = new(http.ServeMux)
	http.Handle("/", http.FileServer(http.Dir(directory)))

	log.Printf("FileServer started at localhost%s with path %s", srv.Addr, directory)

	go func() {
		defer waitGroup.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("FileServer ListenAndServe(): %v", err)
		}
		log.Println("FileServer finished serving")
	}()

	return srv
}

func ShutDown(server *http.Server, waitGroup *sync.WaitGroup) {
	if err := server.Shutdown(context.TODO()); err != nil {
		log.Println("Error shutting down server:", err)
	}
	waitGroup.Wait()
}
