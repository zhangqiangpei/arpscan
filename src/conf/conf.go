package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var mySQLConfig MySQLConfig

type MySQLConfig struct {
	ApiPort  int    `json: "apiPort"`
	Address  string `json: "address"`
	Port     int    `json: "port"`
	DBName   string `json: "dbName"`
	User     string `json: "user"`
	Password string `json: "password"`
	MaxConn  int    `json: "maxConn"`
	MaxIdle  int    `json: "maxIdle"`
}

func GetInst() MySQLConfig {
	return mySQLConfig
}

func InitConfig(filePath string) {
	fmt.Println(filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = json.Unmarshal(data, &mySQLConfig)
	fmt.Println(mySQLConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
