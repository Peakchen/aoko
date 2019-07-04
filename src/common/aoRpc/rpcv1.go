package aoRpc

// add by stefan 20190614 19:49
// add aorpc for between server and server conmunication. 
// server block rpc.
import (
	"sync"
	"context"
	"time"
	"fmt"
	"reflect"
	"container/list"
)

type TAorpcV1 struct {
	models      map[string]interface{}
	wg     		sync.WaitGroup
	ctx 		context.Context
	cancel		context.CancelFunc
	acts 		*list.List
	actchan		chan *TModelActV1
	mutex		sync.Mutex
	retchan     chan *TActRet
}

var Aorpc *TAorpcV1 = nil
func init(){
	Aorpc = &TAorpcV1{}
	Aorpc.Init()
}

func (self *TAorpcV1) Init(){
	self.models = map[string]interface{}{}
	self.acts = list.New()
}

func (self *TAorpcV1) Run(){
	self.ctx, self.cancel = context.WithCancel(context.Background())
	self.wg.Add(2)
	go self.loop()
	go self.loopAct()
}

/*
	take model and func witch func in params,   
*/
func (self *TAorpcV1) Call(key, modelname, funcName string, ins []interface{}, outs []interface{})(error){
	m, ok := self.models[modelname]
	if !ok {
		return fmt.Errorf("can not find model, input model name: %v.", modelname)
	}
	v := reflect.ValueOf(m)
	t := fmt.Sprintf("%s", reflect.TypeOf(m)) 
	f := v.MethodByName(funcName)
	rv := []reflect.Value{}
	for _, in := range ins {
		rv = append(rv, reflect.ValueOf(in))
	}
	//f.Call(rv)
	actkey := key+":"+modelname+":"+funcName
	self.actchan <- &TModelActV1{
		actid:	actkey,
		modf: 	f,
		params: rv,
		mod:	m,
		modt:   t,
	}
	var twg sync.WaitGroup
	twg.Add(1)
	go self.loopRet(actkey, outs, &twg)
	twg.Wait()
	return nil
}

func (self *TAorpcV1) loopRet(actkey string, outs []interface{}, twg *sync.WaitGroup){
	t := time.NewTicker(time.Duration(rpcdealline))
	for {
		select {
		case ar := <-self.retchan:
			if ar.actid == actkey {
				for i, ret := range ar.rets {
					reflect.ValueOf(outs[i]).Set(ret)
				}
				twg.Done()
				return
			}
		case <-t.C:
			// beyond return time, then return nothing.
			twg.Done()
		}
	}
}

func (self *TAorpcV1) loop(){
	defer self.wg.Done()
	//t := time.NewTicker(time.Duration(rpcdealline))
	for {
		select {
		case <-self.ctx.Done():
			self.Exit()
			return
		//case <-t.C:
		
		case act := <-self.actchan:
			if act == nil {
				return
			}
			if self.acts.Len() >= ActChanMaxSize {
				fmt.Println("has enough acts in chan.")
				return
			} 
			self.acts.PushBack(act)
		}
	}
}

func (self *TAorpcV1) loopAct(){
	defer self.wg.Done()
	for {
		if self.acts.Len() == 0 {
			continue
		}
		self.mutex.Lock()
		e := self.acts.Front()
		act := e.Value.(*TModelActV1)
		if act == nil {
			fmt.Println("act value invalid: ", e.Value)
			continue
		}
		mrts := act.modf.Call(act.params)
		self.retchan <- &TActRet{
			actid:	act.actid,
			rets:	mrts,
		}
		self.acts.Remove(e)
		self.mutex.Unlock()
	}
}

func (self *TAorpcV1) Exit(){
	self.cancel()
	self.wg.Wait()
}

func Register(name string, model interface{}) {
	_, ok := Aorpc.models[name]
	if ok {
		return
	}
	Aorpc.models[name] = model
}
