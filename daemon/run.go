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
	fileReader, isExist, err := back.GetFile(fk)
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
		fileReader, isExist, err = fsBackend.GetFile(fk)
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

		// TODO: replace this buf nonsense with io.TeeReader
		// since the file is read twice, need to save it in memory
		buf, err := ioutil.ReadAll(fileReader)
		if err != nil {
			log.WithFields(log.Fields{
				"fk":  fk,
				"err": err,
			}).Info("ioutil.ReadAll(fileReader) failed")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		if err := back.PutFile(fk, bytes.NewReader(buf)); err != nil {
			log.WithFields(log.Fields{
				"fk": fk,
			}).Warn("back.PutFile(fk, buf) failed")
		}

		fileReader = bytes.NewReader(buf)
	}

	if _, err := io.Copy(w, fileReader); err != nil {
		log.WithFields(log.Fields{
			"fk":  fk,
			"err": err,
		}).Warn("io.Copy in sendFile failed")
	}
}

func sendSnap(w http.ResponseWriter, r *http.Request, sk *backend.SnapKey) {
	reader, isExist, err := back.GetSnap(sk)
	if err != nil {
		log.WithFields(log.Fields{
			"sk":  sk,
			"err": err,
		}).Warn("back.GetSnap(sk) failed")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	if !isExist {
		reader, isExist, err = fsBackend.GetSnap(sk)
		if err != nil {
			log.WithFields(log.Fields{
				"sk":  sk,
				"err": err,
			}).Warn("fsBackend.GetSnap(sk) failed")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}
		if !isExist {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Content Not Found"))
			return
		}

		// TODO: replace this buf nonsense with io.TeeReader
		// since the snap is read twice, need to save it in memory
		buf, err := ioutil.ReadAll(reader)
		if err != nil {
			log.WithFields(log.Fields{
				"sk":  sk,
				"err": err,
			}).Info("ioutil.ReadAll(reader) failed")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		ok, err := back.PutSnap(sk, bytes.NewReader(buf))
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

		reader = bytes.NewReader(buf)
	}

	if _, err := io.Copy(w, reader); err != nil {
		log.WithFields(log.Fields{
			"sk":  sk,
			"err": err,
		}).Warn("io.Copy in sendSnap failed")
	}
}

func recvSnap(w http.ResponseWriter, r *http.Request, sk *backend.SnapKey, body io.Reader) {
	if r.Header.Get("X-No-Git") != "" {
		// TODO: replace this "read twice so save" shit with io.TeeReader
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

// parseFileKey expects a path in this format
// <baseWithOne/slash>/<snapshot>/<filepath/which/can/contain/slashes>
// The function returns a *FileKey and a bool indicating whether the
// *FileKey is valid.
func parseFileKey(path string) (*backend.FileKey, bool) {
	elems := strings.Split(path, "/")
	if len(elems) < 4 || elems[len(elems)-1] == "" {
		log.WithFields(log.Fields{
			"path":  path,
			"elems": elems,
		}).Warn("parseFileKey failed")
		return nil, false
	}

	// find the index of the third slash
	thirdSlashIdx, slashCount := 0, 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			slashCount++
			if slashCount >= 3 {
				thirdSlashIdx = i
				break
			}
		}
	}

	return &backend.FileKey{
		Base:     elems[0] + "/" + elems[1],
		Snapshot: elems[2],
		FilePath: path[thirdSlashIdx+1:],
	}, true
}

// parseSnapKey expects a path in this format
// <baseWithOne/slash>/<snapshot>
// The function returns a *SnapKey and a bool indicating whether the
// *SnapKey is valid.
func parseSnapKey(path string) (*backend.SnapKey, bool) {
	elems := strings.Split(path, "/")
	if len(elems) != 3 && !(len(elems) == 4 && elems[3] == "") {
		log.WithFields(log.Fields{
			"path":       path,
			"len(elems)": len(elems),
			"elems":      elems,
		}).Warn("parseSnapKey failed")
		return nil, false
	}
	return &backend.SnapKey{
		Base:     elems[0] + "/" + elems[1],
		Snapshot: elems[2],
	}, true
}

func router(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fileKey, ok := parseFileKey(r.URL.Path[1:])
		if !ok {
			snapKey, ok := parseSnapKey(r.URL.Path[1:])
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("400 Bad Request"))
				return
			}
			sendSnap(w, r, snapKey)
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
