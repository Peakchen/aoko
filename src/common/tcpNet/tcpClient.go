package tcpNet
// client connect server.
import (
	"net"
	"sync"
	//"time"
	"context"
	"fmt"
)

type TcpClient struct {
	wg  	sync.WaitGroup
	ctx 	context.Context
	cancel	context.CancelFunc
	host    string
	s 		*TcpSession
}

func NewClient(host string)*TcpClient{
	return &TcpClient{
		host: host,
	}
}

func (self *TcpClient) Run(){
	c, err := net.Dial("tcp", self.host)
	if err != nil {
		fmt.Println("net dial err: ", err)
		return
	}
	c.(*net.TCPConn).SetNoDelay(true)
	self.wg.Add(1)
	self.ctx, self.cancel = context.WithCancel(context.Background())
	self.s = NewSession(self.host, c, self.ctx)
	self.s.HandleSession()
}

func (self *TcpClient) Send(data []byte){
	self.s.SetSendCache(data)
}

func (self *TcpClient) Exit(){
	self.cancel()
	self.s.exit()
	self.wg.Wait()
}