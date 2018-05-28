package daemon

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Confbase/cfgd/backend"
	"github.com/Confbase/cfgd/backend/fs"
)

var back backend.Backend
var fsBackend *fs.FileSystem

func Run(cfg *Config) {
	b, err := cfg.ToBackend()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	back = b

	fsBackend = fs.New(cfg.FSRootDir)

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
		if _, isFs := back.(*fs.FileSystem); isFs {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Content Not Found"))
			return
		}
		buf, isExist, err = fsBackend.GetFile(fk)
		if err != nil {
			log.WithFields(log.Fields{
				"fk.Base":     fk.Base,
				"fk.Snapshot": fk.Snapshot,
				"fk.FilePath": fk.FilePath,
				"err":         err,
			}).Warn("fsBackend.GetFile(fk) failed")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
		if !isExist {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Content Not Found"))
			return
		}
		if err := back.PutFile(fk, buf); err != nil {
			log.WithFields(log.Fields{
				"fk": fk,
			}).Warn("back.PutFile(fk, buf) failed")
		}
	}

	if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
		log.WithFields(log.Fields{
			"fk.Base":     fk.Base,
			"fk.Snapshot": fk.Snapshot,
			"fk.FilePath": fk.FilePath,
			"err":         err,
		}).Warn("io.Copy in sendFile failed")
	}
}

func recvSnap(w http.ResponseWriter, r *http.Request, sk *backend.SnapKey, body io.Reader) {
	if r.Header.Get("X-No-Git") == "" {
		// since body is read twice, need to save it in memory
		buf, err := ioutil.ReadAll(body)
		if err != nil {
			log.WithFields(log.Fields{
				"sk":  sk,
				"err": err,
			}).Info("ioutil.ReadAll(body) failed")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		ok, err := fsBackend.PutSnap(sk, bytes.NewReader(buf))
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
		// read from beginning of buf in next PutSnap call
		body = bytes.NewReader(buf)
	}
	ok, err := back.PutSnap(sk, body)
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
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("201 Content Created"))
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
