package main

import (
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	Enforce("abac")
}

//func main() {
//	// Initialize a Gorm adapter and use it in a Casbin enforcer:
//	// The adapter will use the MySQL database named "casbin".
//	// If it doesn't exist, the adapter will create it automatically.
//	// You can also use an already existing gorm instance with gormadapter.NewAdapterByDB(gormInstance)
//	a, _ := gormadapter.NewAdapter("mysql", "root:root@tcp(127.0.0.1:3306)/") // Your driver and data source.
//	e, _ := casbin.NewEnforcer("model.conf", a)
//
//	// Or you can use an existing DB "abc" like this:
//	// The adapter will use the table named "casbin_rule".
//	// If it doesn't exist, the adapter will create it automatically.
//	// a := gormadapter.NewAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/abc", true)
//
//	// Load the policy from DB.
//	err := e.LoadPolicy()
//	if err != nil {
//		return
//	}
//	// Check the permission.
//	enforce, err := e.Enforce("alice", "read", "data1")
//	if err != nil {
//		return
//	}
//	fmt.Println(enforce)
//	e.AddPolicy("alice", "read", "data1")
//	// Modify the policy.
//	// e.AddPolicy(...)
//	// e.RemovePolicy(...)
//
//	// Save the policy back to DB.
//	err = e.SavePolicy()
//	if err != nil {
//		return
//	}
//}
