package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"log"
)

// 题目：我们在数据库操作的时候，比如dao层中当遇到一个sql.ErrNoRows的时候，是否应该Wrap这个error，抛给上层？
// 答：不应该Wrap抛给上层，因为sql.ErrNoRows是未查询到数据，应该降级处理

func main() {
	// 业务层调用
	uid := 1
	info, err := getUserInfo(uid)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if info == nil {
		log.Printf("用户#%d,不存在\n", uid)
	} else {
		log.Println(info)
	}
}

func db() (*sql.DB, error) {
	DB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8")
	if err != nil {
		return nil, err
	}

	if err = DB.Ping(); err != nil {
		return nil, err
	}

	return DB, nil
}

// dao
func getUserInfo(uid int) (user map[string]interface{}, err error) {
	DB, err := db()
	if err != nil {
		return nil, errors.Wrap(err, "数据库链接失败")
	}
	defer DB.Close()

	var id int
	var username string
	var age int
	err = DB.QueryRow("select * from user where id=? limit 1", uid).Scan(&id, &username, &age)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "数据查询失败")
	}

	return map[string]interface{}{
		"id":       id,
		"username": username,
	}, nil
}
