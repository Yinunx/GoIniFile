package iniconfig

import (
	"io/ioutil"
	"testing"
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

func TestIniConfig(t *testing.T) {
	//t.Error("hello") //捕获错误信息
	data, err := ioutil.ReadFile("./config.ini")
	if err != nil {
		t.Error("ioutil.Read is wrong")
	}

	var conf Config
	UnMarshal(data, &conf)
	if err != nil {
		t.Error("unmarshal failed, err:", err)
	}
	t.Logf("unmarshal success, conf:%#v, port:%v", conf, conf.ServerConf.Port)
	confData, err := Marshal(conf)
	if err != nil {
		t.Error("unmarshal failed, err:", err)
	}
	t.Logf("marshal success, conf:%s", string(confData))

	//MarshalFile(conf, "C:/loggos/conf")
}

func TestIniConfigFile(t *testing.T) {
	filename := "C:/loggos/conf"
	var conf Config
	conf.ServerConf.Ip = "localhost"
	conf.ServerConf.Port = 88888
	err := MarshalFile(filename, conf)
	if err != nil {
		t.Error("MarshalFile failed, err:", err)
	}

	var conf2 Config
	err = UnMarshalFile(filename, &conf2)
	if err != nil {
		t.Error("unmarshalFile failed, err:", err)
	}
	t.Logf("marshal success, conf:%#v", conf2)
}
