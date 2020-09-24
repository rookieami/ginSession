package ginSession

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"log"
	"strconv"
	"sync"
	"time"
)

//redis版本的session服务

//redisSession redis版本session
type redisSession struct {
	id         string
	data       map[string]interface{}
	modifyFlag bool //修改标志,不修改就不用保存
	expired    int
	rwLock     sync.RWMutex
	client     *redis.Client
}

//NewRedisSession redisSession构造函数
func NewRedisSession(id string, client *redis.Client) (session Session) {
	return &redisSession{
		id:     id,
		data:   make(map[string]interface{}, 8),
		client: client,
	}
}
func (r *redisSession) ID() string {
	return r.id
}

//Load 来自redis中的session 数据
func (r *redisSession) Load() (err error) {
	data, err := r.client.Get(r.id).Bytes()
	if err != nil {
		log.Printf("get session data from redis by %s failed,err:%s\n", r.id, err)
		return
	}
	//unmarshal
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(&r.data)
	if err != nil {
		log.Printf("gob decode session data failed,err:%v\n", err)
		return err
	}
	return err
}

func (r *redisSession) Get(key string) (value interface{}, err error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	value, ok := r.data[key]
	if !ok {
		err = fmt.Errorf("invalid key")
		return
	}
	return
}

func (r *redisSession) Set(key string, value interface{}) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.data[key] = value
	r.modifyFlag = true //修改标志
}

func (r redisSession) Del(key string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	delete(r.data, key)
	r.modifyFlag = true
}

func (r redisSession) Save() {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	if !r.modifyFlag {
		//如果没有修改就不保存
		return
	}
	//修改的话需要再次保存到redis
	//encode
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(r.data)
	if err != nil {
		log.Printf("gob encode r.data failed.err:%v\n", err)
		return
	}
	r.client.Set(r.id, buf.Bytes(), time.Second*time.Duration(r.expired))
	log.Printf("set data %v to redis,\n", buf.Bytes())
	r.modifyFlag = false
}

func (r redisSession) SetExpired(expired int) {
	r.expired = expired
}

type redisSessionMgr struct {
	session map[string]Session
	rwLock  sync.RWMutex
	client  *redis.Client
}

func NewRedisSessionMgr() *redisSessionMgr {
	return &redisSessionMgr{
		session: make(map[string]Session, 1024),
	}

}
func (r *redisSessionMgr) Init(addr string, options ...string) (err error) {
	var (
		password string
		db       int
	)
	if len(options) == 1 {
		password = options[0]
	}
	if len(options) == 2 {
		password = options[0]
		db, err = strconv.Atoi(options[1])
		if err != nil {
			log.Fatalln("invalid redis DB param")
		}
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err = r.client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

//GetSession 加载session并添加到sessionMgr
func (r *redisSessionMgr) GetSession(sessionID string) (sd Session, err error) {
	sd = NewRedisSession(sessionID, r.client)
	err = sd.Load()
	if err != nil {
		return
	}
	r.rwLock.RLock()
	r.session[sessionID] = sd
	r.rwLock.RUnlock()
	return
}

//CreateSession 创建一个新的session
func (r *redisSessionMgr) CreateSession() (sd Session) {
	//生成一个iD
	sessionID := uuid.NewV4().String()
	//创建一个session
	sd = NewRedisSession(sessionID, r.client)
	r.session[sd.ID()] = sd
	return
}

//Clear 在sessionMgr中删除请求session
func (r *redisSessionMgr) Clear(sessionID string) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	delete(r.session, sessionID)
}
