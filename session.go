package ginSession

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

const (
	//cookie中存储的sessionId
	SessionCookieName = "session_id"
	//在gin.Context中的SessionName
	SessionContextName = "session"
)

//Session 具体用户session存储的数据
type Session interface {
	ID() string
	Load() error //加载
	Get(string) (interface{}, error)
	Set(string, interface{})
	Del(string)
	Save()
	SetExpired(int)
}

//SessionMgr session管理者
type SessionMgr interface {
	Init(addr string, options ...string) error //初始化session存储
	GetSession(string) (Session, error)        //通过sessionID得到服务端存储session
	CreateSession() Session                    //创建一个新的session
	Clear(string)                              //清除session数据
}

//Options Cookie 选项
type Options struct {
	Path     string //路径
	Domain   string //域名
	MaxAge   int    //Cookie最大生存时间
	Secure   bool   //是否安全传递->https
	HttpOnly bool   //防止跨域攻击
}

//CreateSessionMgr 通过给定名称数据创建一个SessionMgr
func CreateSessionMgr(name string, addr string, options ...string) (sm SessionMgr, err error) {
	switch name {
	case "memory":
		sm = NewMemSessionMgr() //内存版session管理者
	case "redis":
		sm = NewRedisSessionMgr() //Redis版
	default:
		err = fmt.Errorf("Unsupport%s\n", name) //不支持该类型
		return
	}
	err = sm.Init(addr, options...) //初始化这个SessionMgr
	return
}

//SessionMiddleware gin中间件
func SessionMiddleware(sm SessionMgr, options Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		//为传入请求获取或创建一个session
		//下一个handlerFunc 可以通过c.get得到这个session
		var session Session
		//尝试从cookie中拿到sessionID
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil {
			//得不到sessionID,需要创建一个新的session
			log.Printf("get session_id from Cookie failed,err%s\n", err)
			session = sm.CreateSession()
			sessionID = session.ID()
		} else {
			//可以拿到session
			log.Printf("SessionID :%v\n", sessionID)
			session, err := sm.GetSession(sessionID)
			if err != nil {
				//不能够通过sessionID从服务端存储的session中得到对应session
				log.Printf("get session by %s failed,err:%v\n", sessionID, err)
				session = sm.CreateSession()
				sessionID = session.ID()
			}
		}
		session.SetExpired(options.MaxAge) //设置session过期时间
		c.Set(SessionContextName, session) //将session保存在context中
		//必须在handlerFunc返回之前回写Cookie
		c.SetCookie(SessionContextName, sessionID, options.MaxAge, options.Path, options.Domain, options.Secure, options.HttpOnly)
		defer sm.Clear(sessionID)
		c.Next()
	}
}
