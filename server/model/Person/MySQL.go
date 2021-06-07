package Person

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Person struct {
	Name     string
	Password string
}

const (
	driverName = "mysql"                            //驱动名-这个名字其实就是数据库驱动注册到database/sql 时所使用的名字.
	userName   = "root"                             //用户名
	passWord   = "WW876001"                         //密码
	ip         = "127.0.0.1"                        //ip地址
	port       = "3306"                             //端口号
	dbName     = "tell"                             //数据库名
	ipFormat   = "%s:%s@tcp(%s:%s)/%s?charset=utf8" //格式
)

func DBInit() (*sql.DB, error) {
	DB, err := sql.Open(driverName, fmt.Sprintf(ipFormat, userName, passWord, ip, port, dbName))
	if err != nil {
		return &sql.DB{}, err
	}
	err = DB.Ping()
	if err != nil {
		return &sql.DB{}, err
	}
	return DB, nil
}

// Find 通过name查找返回Person
func Find(DB *sql.DB, name string) (Person, error) {
	var person Person
	rows, err := DB.Query("SELECT `Name`,Password FROM user WHERE binary `Name` = ? ", name)
	if err != nil {
		return person, err
	}
	for rows.Next() {
		err := rows.Scan(&person.Name, &person.Password)
		if err != nil {
			return Person{}, err
		}
		return person, nil
	}
	return person, errors.New("NoFind")
}

// Insert 插入信息
func Insert(DB *sql.DB, person Person) error {
	_, err := DB.Exec("INSERT INTO user (`Name`,Password) VALUES (?,?)", person.Name, person.Password)
	if err != nil {
		return err
	}
	return nil
}

// ReName 修改姓名
func ReName(DB *sql.DB, oldName string, newName string) error {
	_, err := DB.Exec("UPDATE user SET `Name` = ? WHERE binary `Name` = ?", newName, oldName)
	if err != nil {
		return err
	}
	return nil
}

// RePassword 修改密码
func RePassword(DB *sql.DB, Name string, newPassword string) error {
	_, err := DB.Exec("UPDATE user SET password = ? WHERE binary `name` = ?", newPassword, Name)
	if err != nil {
		return err
	}
	return nil
}
