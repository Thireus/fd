package main

import (
	"bytes"
	"context"
	"fmt"
	frida_go "github.com/a97077088/frida-go"
	jsoniter "github.com/json-iterator/go"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type RunParam struct {
	Pid uint
	Name string
	JsPath string
	JsByte bool
	Devi string
}
type Run struct {

}

func (l *Run) Run(param RunParam) error {
	mgr:=frida_go.DeviceManager_Create()
	defer mgr.Close()
	d,err:=ParseDevice(mgr,param.Devi)
	if err!=nil{
		return err
	}
	sysparam,err:=d.QuerySystemParameters()
	if err!=nil{
	    return err
	}
	jssys:=jsoniter.Wrap(sysparam)
	jsos:=jssys.Get("os")
	fmt.Printf("内核平台:%s cpu构架:%s 当前系统:%s(%s)  设备名称:%s 权限:%s \n",jssys.Get("platform").ToString(),jssys.Get("arch").ToString(),jsos.Get(1).Get("id").ToString(),jsos.Get(0).Get("version").ToString(),jssys.Get("name").ToString(),jssys.Get("access").ToString())

	var pid uint
	if param.Pid==0{
		p,err:=d.GetProcessByName(param.Name,frida_go.ProcessMatchOptions{})
		if err!=nil{
			return err
		}
		fmt.Printf("调试进程:%s 进程id:%d 脚本:%s\n",p.Name(),p.Pid(),param.JsPath)
		pid=p.Pid()
	}else{
		pid=param.Pid
		fmt.Printf("进程id:%d 脚本:%s\n",pid,param.JsPath)
	}


	session,err:=d.Attach(pid,frida_go.SessionOptions{})
	if err!=nil{
		return err
	}
	defer session.Detach()
	//fmt.Println("download file example: send({\"type\":\"download\",\"path\":\"test/test.txt\",\"append\":true},new Uint8Array([0x01]).buffer)")


	_,err=os.Stat("./agent/box")
	if err==nil{
		tmplatebox,err:=frida_agent_example.ReadFile("frida-agent-example/agent/box.ts")
		if err!=nil{
			return err
		}
		var boxBuffer bytes.Buffer
		boxBuffer.Write(tmplatebox)
		err=filepath.WalkDir("./agent/box", func(fpath string, d fs.DirEntry, err error) error {
			path:=fpath
			if d.IsDir()==false&&strings.HasPrefix(path,".")==false{
				if strings.HasPrefix(path,"."){
					path=path[1:]
				}
				if strings.HasPrefix(path,"agent\\"){
					path=strings.TrimPrefix(path,"agent\\box\\")
				}
				path=strings.ReplaceAll(path,"\\","/")
				fbyte,err:=ioutil.ReadFile(fpath)
				if err!=nil{
					return err
				}
				ascs:=strings.Builder{}
				for _, b := range fbyte {
					ascs.WriteString(fmt.Sprintf("\\x%02x",b))
				}
				boxBuffer.WriteString(fmt.Sprintf(`Box.MapBox.set("%s","%s")`,path,ascs.String()))
				boxBuffer.WriteByte('\n')
			}
			return nil
		})
		if err!=nil{
			return err
		}
		err=ioutil.WriteFile("./agent/box.ts",boxBuffer.Bytes(),os.ModePerm)
		if err!=nil{
			return err
		}
	}



	fd,err:=ioutil.ReadFile(param.JsPath)
	if err!=nil{
		return err
	}
	fd=append(fd,[]byte("\r\nconsole.log(\"script start\")")...)
	var sc *frida_go.Script
	if param.JsByte{
		sc,err=session.CreateScriptFormBytes(fd,frida_go.ScriptOptions{})
	}else{
		sc,err=session.CreateScript(string(fd),frida_go.ScriptOptions{})
	}
	if err!=nil{
		return err
	}
	ctx,cancel:=context.WithCancel(context.TODO())
	sc.OnDestroyed(func() {
		cancel()
	})


	hslk:=sync.Map{}
	sc.OnMessage(func(sjson jsoniter.Any, data []byte) {
		tp:=sjson.Get("type").ToString()
		if tp=="log"{
			log.Println(sjson.Get("payload").ToString())
		}else if tp=="error"{
			cancel()
			log.Println(sjson.Get("stack").ToString())
			log.Println(sjson.Get("fileName").ToString())
		}else if tp=="send"{
			sendtype:=sjson.Get("payload","type").ToString()
			if sendtype=="download" || sendtype=="down" || sendtype=="downloadfile"{
				fpath:=sjson.Get("payload","path").ToString()
				if fpath==""{
					fpath=sjson.Get("payload","path").ToString()
				}
				appendfile:=sjson.Get("payload","append").ToBool()
				if fpath==""{
					log.Println(sjson.ToString())
					return
				}
				iflk,_:=hslk.LoadOrStore(fpath,sync.Mutex{})
				flk:=iflk.(sync.Mutex)
				flk.Lock()
				defer flk.Unlock()
				toroot:=fmt.Sprintf("./download")
				sdir,sfilename:=filepath.Split(fpath)
				todir:=filepath.Join(toroot,sdir)
				err=os.MkdirAll(todir,os.ModePerm)
				if err!=nil{
				    log.Println(err.Error())
					return
				}

				fg:=os.O_CREATE
				if appendfile==true{
					fg|=os.O_APPEND
				}
				f,err:=os.OpenFile(filepath.Join(todir,sfilename),fg,os.ModePerm)
				if err!=nil{
					log.Println(err.Error())
					return
				}
				defer f.Close()
				_,err=f.Write(data)
				if err!=nil{
					log.Println(err.Error())
					return
				}
			}else {
				log.Println(sjson.ToString())
			}
		}else{
			log.Println(sjson.ToString())
		}
	})
	err=sc.Load()
	if err!=nil{
	    return err
	}
	defer sc.UnLoad()




	<-ctx.Done()
	fmt.Println("脚本运行完毕")
	return nil
}

func NewRun()*Run{
	return &Run{}
}
