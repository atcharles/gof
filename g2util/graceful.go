package g2util

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type (
	//ItfGracefulProcess 接口,提供给关闭后行为
	ItfGracefulProcess interface{ AfterShutdown() }

	//Graceful ...
	Graceful struct {
		Logger LevelLogger `inject:""`

		mu sync.RWMutex
		//信号
		sig chan os.Signal
		//释放资源对象列表
		processList []ItfGracefulProcess
		srvList     []*http.Server
	}
)

// Constructor ...
func (g *Graceful) Constructor() {
	g.sig = make(chan os.Signal)
	g.processList = make([]ItfGracefulProcess, 0)
	g.srvList = make([]*http.Server, 0)
}

// WaitForSignal ...
func (g *Graceful) WaitForSignal() {
	signal.Notify(g.sig)
	g.StartSignal()
}

// RegProcessor ...注册一个对象,用于关闭程序后的行为
func (g *Graceful) RegProcessor(p ItfGracefulProcess) { g.regProcessor(p) }
func (g *Graceful) regProcessor(p ItfGracefulProcess) {
	g.mu.Lock()
	g.processList = append(g.processList, p)
	g.mu.Unlock()
}

// RegHTTPServer ...注册一个http服务
func (g *Graceful) RegHTTPServer(srv *http.Server) { g.regHTTPServer(srv) }
func (g *Graceful) regHTTPServer(srv *http.Server) {
	g.mu.Lock()
	g.srvList = append(g.srvList, srv)
	go func() {
		g.Logger.Println(fmt.Sprintf("HttpServer Listened on:👉 %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			//g.Logger.Errorf("HttpServer witch listening on %s close with error:%s", srv.Addr, err.Error())
			log.Fatalf("HttpServer witch listening on %s close with error:%s\n", srv.Addr, err.Error())
		}
	}()
	g.mu.Unlock()
}

// shutdownAction ...
func (g *Graceful) shutdownAction() {
	g.mu.Lock()
	for _, srv := range g.srvList {
		g.graceShutdownHTTPServer(srv)
	}

	slp := g.processList
	slpIO := make([]ItfGracefulProcess, 0)
	for _, process := range slp {
		//实现了io的对象,等到下一个释放
		if _, ok := process.(io.Writer); !ok {
			TimeoutExecFunc(process.AfterShutdown, time.Second*30)
			continue
		}
		slpIO = append(slpIO, process)
	}

	for _, process := range slpIO {
		TimeoutExecFunc(process.AfterShutdown, time.Second*30)
	}
	g.mu.Unlock()
}

// StartSignal ...监听信号
func (g *Graceful) StartSignal() {
	defer func() {
		signal.Stop(g.sig)
	}()
	for {
		sig := <-g.sig
		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
			g.shutdownAction()
			return
		case syscall.SIGUSR1:
			return
		case syscall.SIGURG, syscall.SIGPIPE, syscall.SIGCHLD:
			//更改https://golang.org/cl/217617提到了此问题：runtime: don't treat SIGURG as a bad signal
		default:
			g.Logger.Infof("unknown Kill Signal:%d; %s\n", sig, sig.String())
		}
	}
}

// graceShutdownHTTPServer ...http 服务优雅关闭
func (g *Graceful) graceShutdownHTTPServer(srv *http.Server) {
	f1 := fmt.Sprintf("[Server listening on->%s]", srv.Addr)
	//Base.StdLogger.Infof("Shutdown %s ...", f1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		g.Logger.Errorf("Shutdown %s error:%s", f1, err.Error())
		return
	}
	g.Logger.Println(fmt.Sprintf("Shutdown: %s exited", f1))
}
