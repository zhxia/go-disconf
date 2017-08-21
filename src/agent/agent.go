package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"log"
	"fmt"
	"reflect"
	"os"
	//"../util"
	"util"
	"encoding/json"
	"sync"
	"strings"
	"flag"
)

var (
	host string
	dir string
)

var usage = `Usage:%s [options]
	Options are:
		-r host 	Connect to remote redis server
		-d directory 	Set config root directory
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
	}
	flag.StringVar(&host, "r", "127.0.0.1:6379", "")
	flag.StringVar(&dir, "d", "/data/config", "")
	flag.Parse()
	ip, err := util.GetLocalIp()
	if err != nil {
		log.Fatalln("get local ip failed,error:", err)
	}
	agent := &Agent{
		Host:host,
		Dir:dir,
		Ip:ip,
	}
	agent.Run()
}

type Agent struct {
	Host string
	Dir  string
	Ip   string
}

func (a *Agent) Run() {
	//由于agent需要同时进行命令接收以及心跳维持，因此需要使用连接池，使用独立的连接，同一个连接无法做到
	pool := &redis.Pool{
		Dial:func() (redis.Conn, error) {
			conn, err := redis.DialTimeout("tcp", a.Host, time.Second * 10, 0, time.Second * 10)
			if err != nil {
				log.Fatalln("connect to redis server failed,error:", err)
			}
			return conn, err
		},
		MaxActive:5,
		MaxIdle:2,
		IdleTimeout:0,
	}
	log.Println("connect to redis server successful!")
	ticker := time.NewTicker(time.Second * 5)
	defer func() {
		defer pool.Close()
		ticker.Stop()
	}()
	go func() {
		//心跳
		conn := pool.Get()
		for t := range ticker.C {
			conn.Do("HSET", util.REDIS_ALIVE_KEY, a.Ip, t.Unix())
		}
	}()
	conn := pool.Get()
	psc := redis.PubSubConn{Conn:conn}
	psc.Subscribe(util.REDIS_MESSAGE_KEY)
	conn.Flush()
	defer conn.Close()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			var message util.Message
			err := json.Unmarshal(v.Data, &message)
			//fmt.Println(message)
			if err != nil {
				panic(err)
			}
			_, err = a.saveData(message)
			if err == nil {
				a.sendResult(pool.Get(), message.Sid, a.Ip)
			}
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case redis.PMessage:
			fmt.Printf("%s:%s %s\n", v.Channel, v.Pattern, v.Data)
		case error:
			fmt.Println("error:", v)
			return
		default:
			fmt.Println(reflect.TypeOf(v))
		}
	}
}

func (a Agent) saveData(message util.Message) (int, error) {
	var rwm sync.RWMutex
	if !strings.HasPrefix(message.Dest, a.Dir) {
		panic("invalid dest ")
	}
	file, err := os.Create(message.Dest)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	rwm.Lock()
	defer rwm.Unlock()
	return file.WriteString(message.Data)
}

func (a Agent) sendResult(conn redis.Conn, sid string, ip string) {
	defer conn.Close()
	key := fmt.Sprintf(util.REDIS_RESULT_KEY, sid)
	conn.Do("HSET", key, ip, time.Now().Unix())
}