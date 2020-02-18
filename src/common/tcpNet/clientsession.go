package tcpNet

// add by stefan

import (
	"strings"
	"sync"
)

var (
	GClient2ServerSession *TClient2ServerSession
)

type TClient2ServerSession struct {
	sync.Mutex

	c2sSession sync.Map
}

func (this *TClient2ServerSession) RemoveSession(key interface{}) {
	this.Lock()
	defer this.Unlock()

	this.c2sSession.Delete(key)
}

func (this *TClient2ServerSession) AddSession(key interface{}, session TcpSession) {
	this.Lock()
	defer this.Unlock()

	this.c2sSession.Store(key, session)
}

func (this *TClient2ServerSession) GetSession(key interface{}) (session TcpSession) {
	this.Lock()
	defer this.Unlock()

	val, exist := this.c2sSession.Load(key)
	if exist {
		session = val.(TcpSession)
	}
	return
}

func (this *TClient2ServerSession) GetSessionByIdentify(key interface{}) (session TcpSession) {
	stridentify, ok := key.(string)
	if ok {
		var (
			dstkey string
		)
		for _, str := range stridentify {
			if str == 32 || str == 0 {
				break
			}
			dstkey += string(str)
		}
		key = strings.TrimSpace(dstkey)
	}
	return this.GetSession(key)
}

func init() {
	GClient2ServerSession = &TClient2ServerSession{}
}
