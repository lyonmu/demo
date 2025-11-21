package interview

import (
	"fmt"
	"sync"
)

// go语言普通`map`和`sync.map`有什么区别和优缺点已经它们的使用场景

func InterviewMap() {

	// 新建
	// 普通 map
	normalMap := make(map[string]int)
	// sync.map
	syncMap := sync.Map{}

	// 插入
	normalMap["key"] = 1
	syncMap.Store("key", 1)

	// 读取
	valueNormal, okNormal := normalMap["key"]
	if !okNormal {
		fmt.Println("key not found")
	}
	fmt.Println(valueNormal)
	valueSync, okSync := syncMap.Load("key")
	if !okSync {
		fmt.Println("key not found")
	}
	fmt.Println(valueSync)
}

type MyError struct {
	Message string
}

func (e *MyError) Error() string {
	return e.Message
}
