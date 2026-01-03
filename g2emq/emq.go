package g2emq

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"

	"github.com/atcharles/gof/v2/g2util"
)

// EmqInstance ..
var EmqInstance *Emq

// Emq ...
type Emq struct {
	Config *g2util.Config     `inject:""`
	Logger g2util.LevelLogger `inject:""`

	opt         *mqtt.ClientOptions
	client      mqtt.Client
	qos         byte
	retained    bool
	wait        time.Duration
	topicPrefix string

	pubC chan func()
}

// prepareOption ...
func (e *Emq) prepareOption() {
	//mqtt.DEBUG = NewLevelLogger("[mqtt]")
	//mqtt.WARN = NewLevelLogger("[mqtt]")
	//mqtt.CRITICAL = NewLevelLogger("[mqtt]")
	//mqtt.ERROR = NewLevelLogger("[mqtt]")

	cfg := e.Config.Viper()
	e.topicPrefix = fmt.Sprintf("%s/", cfg.GetString("name"))
	o1 := mqtt.NewClientOptions()
	o1.AddBroker(strings.Replace(cfg.GetString("emqx.broker"), "{host}", cfg.GetString("global.host"), -1))
	o1.SetStore(mqtt.NewMemoryStore())
	o1.SetConnectRetry(true)
	o1.SetKeepAlive(60 * time.Second)
	o1.SetPingTimeout(2 * time.Second)
	o1.SetClientID(fmt.Sprintf("%s:SYS_%s", cfg.GetString("name"), g2util.ShortUUID()))
	username, password := e.EmqSuperAuth()
	o1.SetUsername(username)
	o1.SetPassword(password)
	o1.SetOnConnectHandler(func(client mqtt.Client) {})
	e.opt = o1
}

// EmqSuperAuth ...
func (e *Emq) EmqSuperAuth() (username, password string) {
	cfg := e.Config.Viper()
	sha256P := sha256.Sum256([]byte(cfg.GetString("emqx.super_password")))
	return cfg.GetString("emqx.super_username"), hex.EncodeToString(sha256P[:])
}

// Client ...
func (e *Emq) Client() mqtt.Client { return e.client }

// Constructor ...
func (e *Emq) Constructor() {
	e.qos = 1
	e.retained = false
	e.wait = time.Second * 2
	e.pubC = make(chan func(), 1000)
	EmqInstance = e
}

// Start ...
func (e *Emq) Start() {
	go func() {
		for {
			f := <-e.pubC
			f()
		}
	}()
}

// PublishC ......
func (e *Emq) PublishC(topic string, payload interface{}) {
	select {
	case e.pubC <- func() {
		if err := e.Publish(topic, payload); err != nil {
			e.Logger.Errorf("[EMQ Publish] [E] topic %s error: %s", topic, err.Error())
		}
	}:
	case <-time.After(time.Millisecond * 100):
		e.Logger.Errorf("[EMQ Publish] [E] publish topic %s timeout", topic)
	}
}

// Publish ...
func (e *Emq) Publish(topic string, payload interface{}) (err error) {
	var buf bytes.Buffer
	switch v := payload.(type) {
	case []byte:
		buf.Write(v)
	case string:
		buf.WriteString(v)
	default:
		buf.WriteString(g2util.JSONDump(v))
	}
	topic = e.topicPrefix + topic
	t := e.client.Publish(topic, e.qos, e.retained, buf.Bytes())
	if !t.WaitTimeout(e.wait) {
		return ErrTokenTimeout
	}
	return t.Error()
}

// Subscribe ...
func (e *Emq) Subscribe(topic string, callback mqtt.MessageHandler) (err error) {
	topic = e.topicPrefix + topic
	t := e.client.Subscribe(topic, e.qos, callback)
	if !t.WaitTimeout(e.wait) {
		return ErrTokenTimeout
	}
	return t.Error()
}

// Dial ...
func (e *Emq) Dial() {
	if err := e.dial(); err != nil {
		log.Fatalln(err)
		return
	}
	e.Logger.Infof("emqx client connected")
	e.Start()
}

// dial ...
func (e *Emq) dial() (err error) {
	e.prepareOption()
	e.client = mqtt.NewClient(e.opt)
	tk := e.client.Connect()
	if !tk.WaitTimeout(e.wait) {
		return ErrTokenTimeout
	}
	return tk.Error()
}

// ErrTokenTimeout declare
var ErrTokenTimeout = errors.New("等待emq服务器Token超时")
