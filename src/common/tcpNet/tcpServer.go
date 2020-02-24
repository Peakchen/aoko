// add by stefan

package tcpNet

import (
	"common/Define"
	"common/Log"
	"common/pprof"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

type TcpServer struct {
	sync.Mutex

	host      string
	pprofAddr string
	listener  *net.TCPListener
	cancel    context.CancelFunc
	cb        MessageCb
	off       chan *SvrTcpSession
	session   *SvrTcpSession
	// person online
	person     int32
	SvrType    Define.ERouteId
	pack       IMessagePack
	SessionMgr IProcessConnSession
	// session id
	SessionID uint64
}

func NewTcpServer(listenAddr, pprofAddr string, SvrType Define.ERouteId, cb MessageCb, sessionMgr IProcessConnSession) *TcpServer {
	return &TcpServer{
		host:       listenAddr,
		pprofAddr:  pprofAddr,
		cb:         cb,
		SvrType:    SvrType,
		SessionMgr: sessionMgr,
		SessionID:  ESessionBeginNum,
		off: 		make(chan *SvrTcpSession, maxOfflineSize),
	}
}

func (this *TcpServer) Run() {
	os.Setenv("GOTRACEBACK", "crash")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", this.host)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	this.listener = listener

	var (
		ctx context.Context
		sw  = sync.WaitGroup{}
	)

	ctx, this.cancel = context.WithCancel(context.Background())
	pprof.Run(ctx)

	this.pack = &ServerProtocol{}
	sw.Add(3)
	go this.loop(ctx, &sw)
	go this.loopoff(ctx, &sw)
	go func() {
		Log.FmtPrintln("[server] run http server, host: ", this.pprofAddr)
		http.ListenAndServe(this.pprofAddr, nil)
	}()
	sw.Wait()
}

func (this *TcpServer) loop(ctx context.Context, sw *sync.WaitGroup) {
	defer func() {
		sw.Done()
		this.Exit(sw)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := this.listener.AcceptTCP()
			if err != nil || c == nil {
				Log.FmtPrintf("can not accept tcp.")
				continue
			}

			c.SetNoDelay(true)
			c.SetKeepAlive(true)
			atomic.AddUint64(&this.SessionID, 1)
			Log.FmtPrintf("[server] accept connect here addr: %v, SessionID: %v.", c.RemoteAddr(), this.SessionID)
			this.session = NewSvrSession(c.RemoteAddr().String(), c, ctx, this.SvrType, this.cb, this.off, this.pack)
			this.session.HandleSession(sw)
			this.online()
		}
	}
}

func (this *TcpServer) loopoff(ctx context.Context, sw *sync.WaitGroup) {
	defer func() {
		sw.Done()
		this.Exit(sw)
	}()
	for {
		select {
		case os, ok := <-this.off:
			if !ok {
				return
			}
			this.offline(os)
		case <-ctx.Done():
			return
		}
	}
}

func (this *TcpServer) online() {
	this.person++
	// rpc notify person online...

}

func (this *TcpServer) offline(os *SvrTcpSession) {
	this.person--
	// rpc notify person offline...

}

func (this *TcpServer) SendMessage() {

}

func (this *TcpServer) Exit(sw *sync.WaitGroup) {
	this.cancel()
	this.listener.Close()
	pprof.Exit()
}

func (this *TcpServer) SessionType() (st ESessionType) {
	return ESessionType_Server
}

func (this *TcpServer) RemoveSession(session *SvrTcpSession) {
	if this.SessionMgr == nil {
		return
	}

	if session.RegPoint != Define.ERouteId_ER_Invalid {
		this.SessionMgr.RemoveSession(session.RemoteAddr)
	} else {
		this.SessionMgr.RemoveSession(session.StrIdentify)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
