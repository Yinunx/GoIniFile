package iniconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

//结构体序列化到文件当中
func MarshalFile(filename string, data interface{}) (err error) {
	result, err := Marshal(data)
	if err != nil {
		return
	}

	return ioutil.WriteFile(filename, result, 0755)
}

//将文件中的反序列化到结构体
func UnMarshalFile(filename string, result interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	return UnMarshal(data, result)
}

//序列化将结构体序列化成配置文件
func Marshal(data interface{}) (result []byte, err error) {
	typeInfo := reflect.TypeOf(data)
	if typeInfo.Kind() != reflect.Struct { //是否结构体
		err = errors.New("please pass struct")
		return
	}

	var conf []string
	valueInfo := reflect.ValueOf(data)
	for i := 0; i < typeInfo.NumField(); i++ {
		sectionField := typeInfo.Field(i) //获取结构体每个字段
		sectionVal := valueInfo.Field(i)  //值
		fieldType := sectionField.Type    //字段类型
		if fieldType.Kind() != reflect.Struct {
			continue
		}

		tagVal := sectionField.Tag.Get("ini") //这是个节
		if len(tagVal) == 0 {
			tagVal = sectionField.Name //没有就把结构体变量中的名字传进去
		}

		section := fmt.Sprintf("\n[%s]\n", tagVal) //节的配置项存起来ServerConf ServerConfig `ini:"server"`
		conf = append(conf, section)

		//再次遍历结构体里面的结构体
		for j := 0; j < fieldType.NumField(); j++ {
			keyField := fieldType.Field(j)
			fieldTagVal := keyField.Tag.Get("ini")
			if len(fieldTagVal) == 0 {
				fieldTagVal = keyField.Name
			}

			valField := sectionVal.Field(j)                                   //拿到当前选项的值
			item := fmt.Sprintf("%s=%v\n", fieldTagVal, valField.Interface()) //格式化成选项
			conf = append(conf, item)
		}
	}

	for _, val := range conf {
		byteVal := []byte(val)
		result = append(result, byteVal...)
	}
	return
}

//反序列化成结构体
func UnMarshal(data []byte, result interface{}) (err error) {

	lineArr := strings.Split(string(data), "\n") //切分成每一行的字符串数组
	/*
		for _, v := range lineArr {
			fmt.Printf("%s\n", v) //必须加\n，不然会输出到缓存里面去
		}
	*/
	typeInfo := reflect.TypeOf(result)
	if typeInfo.Kind() != reflect.Ptr { //是否是指针
		err = errors.New("please pass address")
		return
	}

	typeStruct := typeInfo.Elem()
	if typeStruct.Kind() != reflect.Struct {
		err = errors.New("please pass struct")
		return
	}

	var lastFieldName string
	for index, line := range lineArr {
		line = strings.TrimSpace(line) //字符串前后空格去掉
		if len(line) == 0 {
			continue
		}

		//如果是注释，直接忽略
		if line[0] == ';' || line[0] == '#' {
			continue
		}

		if line[0] == '[' {
			lastFieldName, err = parseSection(line, typeStruct) //解析[xxx]
			if err != nil {
				err = fmt.Errorf("%v lineno:%d", err, index+1)
				return
			}
			continue
		}

		/*
			index := strings.Index(line, "=")
			if index == -1 {
				err = fmt.Errorf("syntax error, line:%s lineno:%d", line, index+1)
				return
			}
		*/
		err = parseItem(lastFieldName, line, result) //解析下面的选项
		if err != nil {
			err = fmt.Errorf("syntax error, line:%s lineno:%d", line, index+1)
			return
		}
	}
	return
}

//解析下面的选项 ip=
func parseItem(lastFieldName string, line string, result interface{}) (err error) {
	index := strings.Index(line, "=")
	if index == -1 {
		err = fmt.Errorf("sytax error, line:%s", line)
		return
	}
	key := strings.TrimSpace(line[0:index])  //前后空格去掉
	val := strings.TrimSpace(line[index+1:]) //配置项里面的value值

	if len(key) == 0 { //key是空的
		err = fmt.Errorf("sytax error, line:%s", line)
		return
	}

	resultValue := reflect.ValueOf(result)                        //指针
	sectionValue := resultValue.Elem().FieldByName(lastFieldName) //选项所对应节，所对应的项  ServerConf

	sectionType := sectionValue.Type()
	if sectionType.Kind() != reflect.Struct {
		err = fmt.Errorf("field:%s must be struct", lastFieldName)
		return
	}

	var keyFieldName string
	for i := 0; i < sectionType.NumField(); i++ { //遍历结构体所有字段,结构体名字
		field := sectionType.Field(i)
		tagVal := field.Tag.Get("ini")
		if tagVal == key {
			keyFieldName = field.Name
			break
		}
	}

	if len(keyFieldName) == 0 { //没有找到
		return
	}

	//fmt.Println(keyFieldName) //结构体变量名字
	//给结构体各变量赋值个赋值
	fieldValue := sectionValue.FieldByName(keyFieldName)
	if fieldValue == reflect.ValueOf(nil) {
		return
	}

	switch fieldValue.Type().Kind() {
	case reflect.String:
		fieldValue.SetString(val) //把配置里面的值映射到结构体中去了
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		intVal, errRet := strconv.ParseInt(val, 10, 64) //转为64位10进制
		if errRet != nil {
			err = errRet
			return
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		intVal, errRet := strconv.ParseUint(val, 10, 64) //转为64位10进制
		if errRet != nil {
			err = errRet
			return
		}
		fieldValue.SetUint(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, errRet := strconv.ParseFloat(val, 64)
		if errRet != nil {
			err = errRet
			return
		}
		fieldValue.SetFloat(floatVal)
	default:
		err = fmt.Errorf("unsupport type:%d", fieldValue.Type().Kind())
	}
	return
}

//解析[xxx]
func parseSection(line string, typeInfo reflect.Type) (fieldName string, err error) {
	//非法[]
	if line[0] == '[' && len(line) <= 2 {
		err = fmt.Errorf("syntax error, invalid section:%s", line) //行号index+1
		return
	}

	//非法[
	if line[0] == '[' && line[len(line)-1] != ']' {
		err = fmt.Errorf("syntax error, invalid section:%s", line) //行号index+1
		return
	}

	if line[0] == '[' && line[len(line)-1] == ']' {
		//非法[ ]
		sectionName := strings.TrimSpace(line[1:(len(line) - 1)])
		if len(sectionName) == 0 {
			err = fmt.Errorf("syntax error, invalid section:%s", line) //行号index+1
			return
		}

		//[server] = sectionName :ie:serve typeInfo2结构体类型信息
		for i := 0; i < typeInfo.NumField(); i++ { //结构体变量个数  tag信息获取
			field := typeInfo.Field(i) //字段
			tagValue := field.Tag.Get("ini")
			if tagValue == sectionName {
				fieldName = field.Name
				break
			}
		}
	}
	return
}
