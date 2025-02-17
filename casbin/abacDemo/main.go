package main

import (
	"abac/user"
	"fmt"
	"github.com/casbin/casbin/v2"
)

func main() {
	user.InitDB()
	// 查询所有用户
	var current user.User
	err := user.DB.Debug().Where("name = ?", "li").First(&current).Error
	if err != nil {
		fmt.Println("err", err)
		return
	}
	// 查询所有数据
	var dataList []user.Data
	err = user.DB.Debug().Find(&dataList).Error
	if err != nil {
		fmt.Println("err", err)
	}
	// 根据用户信息，创建策略
	e, err := casbin.NewEnforcer("./model.conf")
	if err != nil {
		fmt.Println("NewEnforcer failed", err)
		return
	}
	for _, data := range dataList {
		enforce, err := e.Enforce(current.Region, data)
		if err != nil {
			fmt.Println("Enforce failed", err)
			return
		}
		fmt.Println(enforce, data)
	}
}
