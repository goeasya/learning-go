package provider

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"time"
)

type RedisProvider struct {
	client *redis.Client
}

func (p *RedisProvider) SessionInit(sid string) (Session, error) {
	var val = make(map[interface{}]interface{}, 0)
	val["timeAccessed"] = time.Now()
	p.client.Set(sid, val, time.Second*3600)

	return &RedisSessionStore{
		client: p.client,
		sid:    sid,
	}, nil
}

func (p *RedisProvider) SessionRead(sid string) (Session, error) {
	val := p.client.Get(sid)
	if val.Err() != nil {
		return p.SessionInit(sid)
	} else {
		return &RedisSessionStore{
			client: p.client,
			sid:    sid,
		}, nil
	}
}

func (p *RedisProvider) SessionDestroy(sid string) error {
	return p.client.Del(sid).Err()
}

func (p *RedisProvider) SessionGC(maxLifeTime int64) {
	return
}

type RedisSessionStore struct {
	client *redis.Client
	sid    string
}

func (s *RedisSessionStore) Set(key, value interface{}) error {
	var val map[interface{}]interface{}
	v := s.client.Get(s.sid)
	if v.Err() != nil {
		return v.Err()
	}
	err := json.Unmarshal([]byte(v.Val()), &val)
	if err != nil {
		return nil
	}

	val[key] = value
	val["timeAccessed"] = time.Now()

	s.client.Set(s.sid, val, time.Second*3600)
	return nil
}

func (s *RedisSessionStore) Get(key interface{}) (interface{}, error) {
	var value map[interface{}]interface{}
	v := s.client.Get(s.sid)
	if v.Err() != nil {
		return nil, v.Err()
	}

	err := json.Unmarshal([]byte(v.Val()), &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *RedisSessionStore) Delete(key interface{}) error {
	var value map[interface{}]interface{}
	v := s.client.Get(s.sid)
	if v.Err() != nil {
		return v.Err()
	}
	err := json.Unmarshal([]byte(v.Val()), &value)
	if err != nil {
		return nil
	}

	delete(value, key)
	s.client.Set(s.sid, value, time.Second*3600)
	return nil
}

func (s *RedisSessionStore) SessionId() string {
	return s.sid
}

func NewRedisProvider(addr, password string) Provider {
	opt := redis.Options{
		Addr:     addr,
		Password: password,
	}
	cli := redis.NewClient(&opt)
	return &RedisProvider{
		client: cli,
	}
}
