package conf

import (
	"encoding/json"
	"github.com/name5566/leaf/log"
	"io/ioutil"
)

var Server struct {
	LogLevel    string
	LogPath     string
	WSAddr      string
	CertFile    string
	KeyFile     string
	TCPAddr     string
	HTTPPort    string
	MaxConnNum  int
	ConsolePort int
	ProfilePath string

	MongoDBAddr string
	MongoDBAuth string
	MongoDBUser string
	MongoDBPwd  string

	TokenServer      string
	CenterServer     string
	CenterServerPort int
	DevName          string
	GameID           string
	Dev_dezhoupoker  string
	CenterUrl        string

	LogAddr string
}

func init() {
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v", err)
	}
}
