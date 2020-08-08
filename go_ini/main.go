package main

import (
	"fmt"
	"go_ini/iniconfig"
)

type Config struct {
	ServerConf ServerConfig `ini:"server"`
	MySqlConf  MysqlConfig  `ini:"mysql"`
}

type ServerConfig struct {
	Ip   string `ini:"ip"`
	Port int    `ini:"port"`
}

type MysqlConfig struct {
	Username string  `ini:"username"`
	Passwd   string  `ini:"passwd"`
	Database string  `ini:"database"`
	Host     string  `ini:"host"`
	Port     uint    `ini:"port"`
	Timeout  float64 `ini:"timeout"`
}

func main() {
	filename := "C:/goland/code/src/go_ini/iniconfig/config.ini"
	var conf Config
	err := iniconfig.UnMarshalFile(filename, &conf)
	if err != nil {
		fmt.Println("unmarshal faild, err:", err)
	}
	fmt.Printf("conf:%#v\n", conf)
}
