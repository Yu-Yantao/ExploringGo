package user

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User 用户
type User struct {
	ID     uint   `gorm:"primaryKey,autoIncrement"` // 主键
	Name   string `json:"name"`                     // 姓名
	Region string `json:"region"`                   // 区域
}

type Data struct {
	ID         uint   `gorm:"primaryKey,autoIncrement"` // 主键
	State      string `json:"state"`                    // 州
	National   string `json:"national"`                 // 国家
	Provincial string `json:"provincial"`               // 省份
	City       string `json:"city"`                     // 城市
	Area       string `json:"area"`                     // 地区
	Info       string `json:"info"`                     // 数据
}

func InitDB() {
	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Init DB error", err)
	}
	fmt.Println("Init DB success")
	DB = db
}
func AutoMigrate() {
	if DB != nil {
		err := DB.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(&User{}, &Data{})
		if err != nil {
			fmt.Println("AutoMigrate error", err)
		}
		fmt.Println("AutoMigrate success")
	}
}
