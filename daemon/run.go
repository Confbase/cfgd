package daemon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Confbase/cfgd/backend"
)

var back backend.Backend

func sendFile(w http.ResponseWriter, r *http.Request, fk *backend.FileKey) {
	buf, isExist, err := back.GetFile(fk)
	if err != nil {
		log.WithFields(log.Fields{
			"fk.Base":     fk.Base,
			"fk.Snapshot": fk.Snapshot,
			"fk.FilePath": fk.FilePath,
			"err":         err,
		}).Warn("back.GetFile(fk) failed")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	if !isExist {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Content Not Found"))
		return
	}
	totalWritten := 0
	for totalWritten < len(buf) {
		bytesWritten, err := w.Write(buf[totalWritten:])
		if err != nil {
			log.WithFields(log.Fields{
				"fk.Base":     fk.Base,
				"fk.Snapshot": fk.Snapshot,
				"fk.FilePath": fk.FilePath,
				"err":         err,
			}).Warn("w.Write(buf) failed")
			return
		}
		totalWritten += bytesWritten
	}
}

func recvSnap(w http.ResponseWriter, r *http.Request, sk *backend.SnapKey, snapReader io.Reader) {
	ok, err := back.PutSnap(sk, snapReader)
	if err != nil {
		log.WithFields(log.Fields{
			"sk":  sk,
			"err": err,
		}).Info("500 Internal Server Error")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	w.Write([]byte("200 OK"))
}

func parseFileKey(path string) (*backend.FileKey, bool) {
	elems := strings.Split(path, "/")
	if len(elems) < 3 {
		return nil, false
	}

	firstSlash := strings.Index(path, "/")
	secondSlash := strings.Index(path[firstSlash+1:], "/")

	return &backend.FileKey{
		Base:     elems[0],
		Snapshot: elems[1],
		FilePath: path[firstSlash+secondSlash+2:],
	}, true
}

func parseSnapKey(path string) (*backend.SnapKey, bool) {
	elems := strings.Split(path, "/")
	if len(elems) != 2 {
		return nil, false
	}
	return &backend.SnapKey{
		Base:     elems[0],
		Snapshot: elems[1],
	}, true
}

func router(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fileKey, ok := parseFileKey(r.URL.Path[1:])
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad Request"))
			return
		}
		sendFile(w, r, fileKey)
		return
	} else if r.Method == http.MethodPost {
		snapKey, ok := parseSnapKey(r.URL.Path[1:])
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad Request"))
			return
		}
		recvSnap(w, r, snapKey, r.Body)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 Bad Request"))
	return
}

func Run(cfg *Config) {
	b, err := cfg.ToBackend()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	back = b

	log.WithFields(log.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
	}).Info("launching daemon")

	http.HandleFunc("/", router)

	addr := fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
