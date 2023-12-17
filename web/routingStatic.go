package web

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func (s webServer) initStaticRoutes() {
	s.router.PathPrefix("").HandlerFunc(handleStatic)
}

const staticPath = "public"
const indexPath = "index.html"

func handleStatic(w http.ResponseWriter, r *http.Request) {

	log.Infof("handleStatic %s", r.URL)

	var workDir string
	var path string
	var err error

	workDir, err = os.Getwd()

	if runtime.GOOS == "windows" {
		path = filepath.Join(workDir, path, staticPath)
	} else {
		path, err = filepath.Abs(r.URL.Path)

		// Forbid directory browsing.
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		path = filepath.Join(path, staticPath)
	}

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		// Serve index.html.
		http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
		return
	} else if err != nil {
		// An error occurred.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve files.
	http.FileServer(http.Dir(staticPath)).ServeHTTP(w, r)
}
