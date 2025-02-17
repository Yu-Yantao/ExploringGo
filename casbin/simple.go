package main

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"log"
)

func check(e *casbin.Enforcer, sub, act, obj, dom string, withDomain bool) {
	var ok bool
	if withDomain {
		ok, _ = e.Enforce(sub, act, obj, dom)
	} else {
		ok, _ = e.Enforce(sub, act, obj)
	}
	fmt.Println(ok)
}

func Enforce(model string) {
	var e *casbin.Enforcer
	var err error
	// 加载模型和策略
	switch model {
	case "acl":
		e, err = casbin.NewEnforcer("./acl/model.conf", "./acl/policy.csv")
	case "rbac":
		e, err = casbin.NewEnforcer("./rbac/model.conf", "./rbac/policy.csv")
	case "rbacWithDomain":
		e, err = casbin.NewEnforcer("./rbacWithDomain/model.conf", "./rbacWithDomain/policy.csv")
	case "abac":
		e, err = casbin.NewEnforcer("./abac/model.conf")
	default:
		fmt.Println("invalid model")
	}
	if err != nil {
		log.Fatalf("NewEnforecer failed:%v\n", err)
	}
	// 测试
	var ok bool
	switch model {
	case "acl", "rbac":
		ok, _ = e.Enforce("alice", "read", "data1")
		fmt.Println(ok)
		ok, _ = e.Enforce("alice", "write", "data1")
		fmt.Println(ok)
	case "rbacWithDomain":
		ok, _ = e.Enforce("alice", "read", "data1", "tenant1")
		fmt.Println(ok)
		ok, _ = e.Enforce("alice", "write", "data1", "tenant1")
		fmt.Println(ok)
	case "abac":
		// abac
		ok, _ = e.Enforce("alice", "read", Resource{Name: "data1", Owner: "alice"})
		fmt.Println(ok)
	}

}

type Resource struct {
	Name  string
	Owner string
}
