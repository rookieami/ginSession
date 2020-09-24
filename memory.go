package ginSession

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"sync"
)

//memory 版本session 服务
//memSession session保存在内存中
type memSession struct {
	id      string
	data    map[string]interface{}
	expired int
	rwLock  sync.RWMutex
}

//NewMemSession 构造函数
func NewMemSession(id string) *memSession {
	return &memSession{
		id:   id,
		data: make(map[string]interface{}, 8),
	}

}

func (m *memSession) ID() string {
	return m.id
}

func (m *memSession) Load() (err error) {
	return
}
func (m *memSession) Get(key string) (value interface{}, err error) {
	m.rwLock.RLock() //加读锁
	defer m.rwLock.RUnlock()
	value, ok := m.data[key]
	if !ok {
		err = fmt.Errorf("invalid key")
		return
	}
	return
}
func (m *memSession) Set(key string, value interface{}) {
	m.rwLock.Lock() //写锁
	defer m.rwLock.Unlock()
	m.data[key] = value
}
func (m *memSession) Del(key string) {
	m.rwLock.Lock() //写锁
	defer m.rwLock.Unlock()
	delete(m.data, key)
}

//Save 保存session数据
func (m *memSession) Save() {
	return
}

//SetExpired 设置过期时间
func (m *memSession) SetExpired(expired int) {
	m.expired = expired
}

//MemSession 内存版本管理器
type MemSessionMgr struct {
	session map[string]Session
	rwLock  sync.RWMutex
}

//NewMemSessionMgr 构造函数
func NewMemSessionMgr() *MemSessionMgr {
	return &MemSessionMgr{
		session: make(map[string]Session, 1024),
	}
}

//初始化
func (m *MemSessionMgr) Init(addr string, options ...string) (err error) {
	return
}

//GetSession 通过sessionID从后台得到对应session
func (m *MemSessionMgr) GetSession(sessionID string) (sd Session, err error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	sd, ok := m.session[sessionID]
	if !ok {
		err = fmt.Errorf("invalid session id")
		return
	}
	return
}

//CreateSession	创建一个新的session
func (m *MemSessionMgr) CreateSession() (sd Session) {
	//生成一个sessionID
	sessionID := uuid.NewV4()
	//创建一个新的内存版session
	sd = NewMemSession(sessionID.String())
	m.session[sd.ID()] = sd
	return
}

//Clear 在管理器中删除请求session
func (m *MemSessionMgr) Clear(sessionID string) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	delete(m.session, sessionID)
}
