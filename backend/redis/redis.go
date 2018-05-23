package redis

import (
	"bufio"
	"fmt"
	"io"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"

	"github.com/Confbase/cfgd/backend"
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

func (rb *RedisBackend) GetFile(fk *backend.FileKey) ([]byte, bool, error) {
	redisKey := fmt.Sprintf("%v/%v", fk.Base, fk.Snapshot)
	out, err := rb.client.HGet(redisKey, fk.FilePath).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("HGET failed: %v", err)
	}
	return out, true, nil
}

func (rb *RedisBackend) PutFile(fk *backend.FileKey, buf []byte) error {
	redisKey := fmt.Sprintf("%v/%v", fk.Base, fk.Snapshot)
	_, err := rb.client.HSet(redisKey, fk.FilePath, buf).Result()
	if err != nil {
		log.WithFields(log.Fields{
			"redisKey":    redisKey,
			"fk.FilePath": fk.FilePath,
		}).Warn("HSET failed")
		return fmt.Errorf("HMSET failed: %v", err)
	}
	return nil
}

func (rb *RedisBackend) PutSnap(sk *backend.SnapKey, r io.Reader) (bool, error) {
	firstFour := make([]byte, 4)
	if _, err := io.ReadFull(r, firstFour); err != nil {
		if err == io.EOF {
			log.WithFields(log.Fields{
				"sk": sk,
			}).Info("failed to read first four bytes")
			return false, nil
		}
		return false, fmt.Errorf("read failed: %v", err)
	}
	if string(firstFour) != "PUT " {
		log.WithFields(log.Fields{
			"sk": sk,
		}).Info("first four bytes weren't 'PUT '")
		return false, nil
	}

	br := bufio.NewReader(r)
	line, err := br.ReadSlice('\n')
	if err != nil {
		if err == io.EOF {
			log.WithFields(log.Fields{
				"line": string(line),
				"sk":   sk,
			}).Info("failed to find '\\n' before EOF")
			return false, nil
		}
		return false, fmt.Errorf("br.ReadString('\n') failed: %v", err)
	}
	payloadRedisKey := string(line[:len(line)-1])
	redisKey := fmt.Sprintf("%v/%v", sk.Base, sk.Snapshot)
	if payloadRedisKey != redisKey {
		log.WithFields(log.Fields{
			"payloadRedisKey": payloadRedisKey,
			"redisKey":        redisKey,
			"sk":              sk,
		}).Info("sk key from URL path does not match payload key")
		return false, nil
	}

	pipe := rb.client.TxPipeline()

	if _, err := pipe.Del(redisKey).Result(); err != nil {
		return false, fmt.Errorf("HDEL failed: %v", err)
	}

	snapReader := snapshot.NewReader(br)
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
