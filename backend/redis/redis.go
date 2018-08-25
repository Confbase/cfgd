package redis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"

	"github.com/Confbase/cfgd/backend"
	"github.com/Confbase/cfgd/cfgsnap/snapmsg"
	"github.com/Confbase/cfgd/snapshot"
)

type RedisBackend struct {
	client *redis.Client
}

func New(host, port string) *RedisBackend {
	return &RedisBackend{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", host, port),
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
}

func (rb *RedisBackend) GetFile(fk *backend.FileKey) (io.Reader, bool, error) {
	redisKey := fmt.Sprintf("%v/%v", fk.Base, fk.Snapshot)
	out, err := rb.client.HGet(redisKey, fk.FilePath).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("HGET failed: %v", err)
	}
	return bytes.NewReader(out), true, nil
}

func (rb *RedisBackend) PutFile(fk *backend.FileKey, r io.Reader) error {
	redisKey := fmt.Sprintf("%v/%v", fk.Base, fk.Snapshot)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		log.WithFields(log.Fields{
			"redisKey":    redisKey,
			"fk.FilePath": fk.FilePath,
		}).Warn("ioutil.ReadAll(r) failed")
		return err
	}
	if _, err := rb.client.HSet(redisKey, fk.FilePath, buf).Result(); err != nil {
		log.WithFields(log.Fields{
			"redisKey":    redisKey,
			"fk.FilePath": fk.FilePath,
		}).Warn("HSET failed")
		return fmt.Errorf("HMSET failed: %v", err)
	}
	return nil
}

func (rb *RedisBackend) GetSnap(sk *backend.SnapKey) (io.Reader, bool, error) {
	r, w := io.Pipe()
	go func() {
		redisKey := fmt.Sprintf("%v/%v", sk.Base, sk.Snapshot)
		kvs, err := rb.client.HGetAll(redisKey).Result()
		if err != nil && err != redis.Nil {
			log.WithFields(log.Fields{
				"err": err,
				"sk":  sk,
			}).Warn("in GetSnap, rb.client.HGetAll failed")
			w.Close()
			return
		}
		for filePath, contents := range kvs {
			if err := snapmsg.WriteString(w, filePath, contents); err != nil {
				log.WithFields(log.Fields{
					"err":      err,
					"sk":       sk,
					"filePath": w,
				}).Warn("in GetSnap, build.WriteString failed")
				w.Close()
				return
			}
		}
		w.Close()
	}()
	header := fmt.Sprintf("PUT %v\n", sk.ToHeaderKey())
	return io.MultiReader(strings.NewReader(header), r), true, nil
}

func (rb *RedisBackend) PutSnap(sk *backend.SnapKey, r io.Reader) (bool, error) {
	snapReader := snapshot.NewReader(bufio.NewReader(r))
	redisKey := sk.ToHeaderKey()
	if isOk, err := snapReader.VerifyHeader(redisKey); err != nil {
		return false, err
	} else if !isOk {
		return false, nil
	}

	pipe := rb.client.TxPipeline()
	if _, err := pipe.Del(redisKey).Result(); err != nil {
		return false, fmt.Errorf("HDEL failed: %v", err)
	}

	for {
		sf, done, err := snapReader.Next()
		if err != nil {
			return false, fmt.Errorf("snapReader failed: %v", err)
		}
		if done {
			break
		}
		_, err = pipe.HSet(redisKey, string(sf.FilePath), sf.Body).Result()
		if err != nil {
			return false, fmt.Errorf("HSET failed: %v", err)
		}
	}

	cmdErrs, err := pipe.Exec()
	if err != nil {
		return false, fmt.Errorf("pipe.Exec() failed: %v", err)
	}
	for _, cmdErr := range cmdErrs {
		if err := cmdErr.Err(); err != nil {
			return false, fmt.Errorf("redis cmdErr: %v", err)
		}
	}

	return true, nil
}
