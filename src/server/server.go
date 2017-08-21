package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"log"
	//"../util"
	"util"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"net/http"
)

var (
	host string
)
var usage = `Usage:%s [options]
	Options are:
		-h host 	Connect to remote redis server
		-d directory 	Set config root directory
`
/**
curl -X post --data '{"sid":"8113197b-c4cc-4cf8-830e-8257bbc8b59d","dest":"/data/config/service.conf","data":"{\"servers\":[{\"host\":\"127.0.0.1\",\"port\":10010,\"weigth\":20,\"status\":\"online\"}]}"}' http://localhost:8487/sendMessage
 */
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
	}
	flag.StringVar(&host, "h", "127.0.0.1:6379", "")
	flag.Parse()
	server := &Server{
		host:host,
	}
	server.Run()
}

type Server struct {
	host string
}

func (s *Server) Run() {
	http.HandleFunc("/sendMessage", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		var msg util.Message
		err := decoder.Decode(&msg)
		if err != nil {
			panic(err)
		}
		log.Println("Message:", msg)
		s.sendCommand(msg)
	})
	log.Println("Server is Running...")
	err := http.ListenAndServe(":8487", nil)
	if err != nil {
		log.Fatalln("ListenAndServe:", err)
	}
}

func (s *Server) sendCommand(msg util.Message) {
	conn, err := redis.DialTimeout("tcp", s.host, time.Second * 10, time.Second * 5, time.Second * 5)
	if err != nil {
		log.Fatalln("connect to redis server failed,error:", err)
	}
	defer conn.Close()
	log.Println("connect to redis server successful!")
	/*uuid := uuid.Must(uuid.NewRandom()).String()
	message := &util.Message{
		Sid:uuid,
		Dest:"/data/config/service.conf",
		Data:`{"servers":[{"host":"127.0.0.1","port":10010,"weigth":20,"status":"online"}]}`,
	}*/
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("message encode error:", err)
	}
	conn.Send("PUBLISH", util.REDIS_MESSAGE_KEY, data)
	conn.Flush()
}