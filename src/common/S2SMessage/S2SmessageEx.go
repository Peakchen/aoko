/*
Copyright (c) <year> <copyright holders>

"Anti 996" License Version 1.0 (Draft)

Permission is hereby granted to any individual or legal entity
obtaining a copy of this licensed work (including the source code,
documentation and/or related items, hereinafter collectively referred
to as the "licensed work"), free of charge, to deal with the licensed
work for any purpose, including without limitation, the rights to use,
reproduce, modify, prepare derivative works of, distribute, publish 
and sublicense the licensed work, subject to the following conditions:

1. The individual or the legal entity must conspicuously display,
without modification, this License and the notice on each redistributed 
or derivative copy of the Licensed Work.

2. The individual or the legal entity must strictly comply with all
applicable laws, regulations, rules and standards of the jurisdiction
relating to labor and employment where the individual is physically
located or where the individual was born or naturalized; or where the
legal entity is registered or is operating (whichever is stricter). In
case that the jurisdiction has no such laws, regulations, rules and
standards or its laws, regulations, rules and standards are
unenforceable, the individual or the legal entity are required to
comply with Core International Labor Standards.

3. The individual or the legal entity shall not induce, metaphor or force
its employee(s), whether full-time or part-time, or its independent
contractor(s), in any methods, to agree in oral or written form, to
directly or indirectly restrict, weaken or relinquish his or her
rights or remedies under such laws, regulations, rules and standards
relating to labor and employment as mentioned above, no matter whether
such written or oral agreement are enforceable under the laws of the
said jurisdiction, nor shall such individual or the legal entity
limit, in any methods, the rights of its employee(s) or independent
contractor(s) from reporting or complaining to the copyright holder or
relevant authorities monitoring the compliance of the license about
its violation(s) of the said license.

THE LICENSED WORK IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE COPYRIGHT HOLDER BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN ANY WAY CONNECTION WITH THE
LICENSED WORK OR THE USE OR OTHER DEALINGS IN THE LICENSED WORK.
*/

package S2SMessage

import (
	"github.com/gorilla/websocket"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	//"strconv"
	. "common/Define"
)

type S2SDispatchMHandler func(msg []byte, c net.Conn)
type C2SRoutehandler func(msg *SS_MsgRoute_Req, c* websocket.Conn)

var(
	messageHandler  map[int32] C2SRoutehandler
	s2sdispatchMsgMap  map[int32] S2SDispatchMHandler 
)

func S2SRouteRegister(id int32, handler C2SRoutehandler){
	messageHandler[id] = handler
}

func S2SDispatchMRegister(id int32, handler S2SDispatchMHandler){
	s2sdispatchMsgMap[id] = handler
}

func DispatchClientMessage(msg []byte, c* websocket.Conn){

	var msg_base = &SS_MsgRoute_Req{}
	var pm = proto.Unmarshal(msg, msg_base)
	if pm == nil {
		log.Fatal("unmarshal message fail.")
		return
	}

	cb, ok := messageHandler[msg_base.Operid]
	if ok {
		cb(msg_base, c)
	}
}
// message dispatch route.
func DispatchMessage(msg []byte, c net.Conn, srcSvr, dstSvr int32){
	// var msg_base = &C2SMsgRoute{}
	// var pm = proto.Unmarshal(msg, msg_base)
	// if pm == nil {
	// 	log.Fatal("unmarshal message fail.")
	// 	return
	// }
	// cb, ok := messageHandler[*msg_base.Operid]
	// if ok {
	// 	cb(msg_base, c)
	// }

	if srcSvr == dstSvr {
		log.Fatal("source and destination serverid is equal, id: ", dstSvr)
		return
	}
	if srcSvr == int32(ERouteId_ER_Invalid){
		log.Fatal("source serverid is invalid.",)
		return
	}
	if dstSvr == int32(ERouteId_ER_Invalid){
		log.Fatal("source serverid is invalid.",)
		return
	}
	if _, ok := ERouteId_name[srcSvr]; !ok {
		log.Fatal("can not find source serverid: ", srcSvr)
		return
	}
	if _, ok := ERouteId_name[dstSvr]; !ok {
		log.Fatal("can not find destination serverid: ", dstSvr)
		return
	}
	//strconv.Itoa(int(srcSvr))+strconv.Itoa(int(dstSvr))
	
}

func PostMessage(pb proto.Message, c net.Conn){
	msg, err := proto.Marshal(pb)
	if err == nil {
		log.Fatal("Marshal message fail.")
		return
	}

	_, err = c.Write(msg)
	if err != nil {
		log.Fatal("Write close.")
		return
	}
}