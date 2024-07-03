package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils"

	"github.com/kardianos/service"
	"github.com/natefinch/lumberjack"
)

const (
	AppName    = "HttpServer"
	AppVersion = "0.0.1"
	AppConf    = "/usr/local/httpserver/httpserver.json"
)

var (
	h, v, d, u        bool
	c                 string
	conf              Config
	SvrHttp, SvrHttps *http.Server

	//SSL/TLS协议信息泄露漏洞(CVE-2016-2183)
	//解决方法 避免使用DES算法
	//MinVersion: tls.VersionTLS13,
	csArr = []uint16{
		//tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		//tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		//tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_FALLBACK_SCSV}
)

type Config struct {
	HttpPort          string
	HttpsPort         string
	HttpsUse          bool
	HttpsCrt          string
	HttpsKey          string
	LogFilePath       string
	LogFileSize       int
	LogFileNum        int
	LogSaveDay        int
	AllowOrigin       string
	ReadHeaderTimeout int // 用于限制服务器读取请求头的最大时间, 防止慢速HTTP头攻击
}

func InitConf(file string) {
	d, err := utils.ReadAllFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(d, &conf)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(conf)
}

func InitLog(file string) {
	//log.SetFlags(log.LstdFlags | log.Lshortfile) // 前台打印
	//return
	l := new(lumberjack.Logger)
	l.Filename = file
	l.MaxSize = conf.LogFileSize   // megabytes
	l.MaxBackups = conf.LogFileNum // number
	l.MaxAge = conf.LogSaveDay     // days

	log.SetOutput(l)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("==================================================")
	log.Println("==", AppName, "Version:", AppVersion)
	log.Println("== StartTime:", utils.GetYMDHMS())
	log.Println("==================================================")
	log.Println(h, v, d, u, c)
	log.Printf("%#v", conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for {
			<-c
			l.Rotate()
		}
	}()
}

////////////////////////////////////////////////////////////
// 守护进程 且 注册为系统服务(开机启动)
////////////////////////////////////////////////////////////
type program struct{}

func (p *program) run() {
	InitConf(c)
	InitLog(conf.LogFilePath)

	go ReportIDs()

	http.HandleFunc("/", HttpServer)

	log.Printf("start http listen on :%s", conf.HttpPort)
	go func() {
		SvrHttp = &http.Server{
			Addr:              fmt.Sprintf(":%s", conf.HttpPort),
			ReadHeaderTimeout: time.Duration(conf.ReadHeaderTimeout) * time.Second,
		}
		err := SvrHttp.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if conf.HttpsUse == true {
		log.Printf("start https listen on :%s", conf.HttpsPort)
		/*
			go func() {
				SvrHttps = &http.Server{
					Addr:              conf.HttpsListen,
					ReadHeaderTimeout: time.Duration(conf.ReadHeaderTimeout) * time.Second,
					TLSConfig: &tls.Config{
						CipherSuites: csArr,
					},
				}
				err := SvrHttps.ListenAndServeTLS(conf.HttpsCrt, conf.HttpsKey)
				if err != nil {
					log.Fatal(err)
				}
			}()
		*/
	}

	select {}
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	flag.BoolVar(&h, "h", false, "print help")
	flag.BoolVar(&v, "v", false, "print version")
	flag.BoolVar(&d, "d", false, "run in deamon")
	flag.BoolVar(&u, "u", false, "stop in deamon")
	flag.StringVar(&c, "c", AppConf, "config file")
	flag.Parse()
	//flag.Usage()
	log.Println(h, v, d, u, c)
	if h {
		flag.PrintDefaults()
		return
	}
	if v {
		log.Println(AppVersion)
		return
	}

	sc := new(service.Config)
	sc.Name = AppName
	sc.DisplayName = AppName
	sc.Description = AppName

	prg := new(program)
	s, err := service.New(prg, sc)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if u {
		err = service.Control(s, "stop")
		if err != nil {
			log.Println(err)
		} else {
			log.Println("service stopped")
		}
		err = service.Control(s, "uninstall")
		if err != nil {
			log.Println(err)
		} else {
			log.Println("service uninstalled")
		}
		return
	}

	if !d {
		prg.run()
		return
	}

	err = service.Control(s, "stop")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("service stopped")
	}
	err = service.Control(s, "uninstall")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("service uninstalled")
	}
	err = service.Control(s, "install")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("service installed")
	}
	err = service.Control(s, "start")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("service started")
	}
}
