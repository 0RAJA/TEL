package Person

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
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

func Md5Str(bt [md5.Size]byte) string {
	result := ""
	for _, i := range bt {
		result += string(i)
	}
	return strings.ToLower(result)
}
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
	hashName := Md5Str(md5.Sum([]byte(name)))
	var person Person
	rows, err := DB.Query("SELECT `Name`,Password FROM user WHERE `Name` = ? ", hashName)
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
	hashName, hashPassword := Md5Str(md5.Sum([]byte(person.Name))), Md5Str(md5.Sum([]byte(person.Password)))
	_, err := DB.Exec("INSERT INTO user (`Name`,Password) VALUES (?,?)", hashName, hashPassword)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// ReName 修改姓名
func ReName(DB *sql.DB, oldName string, newName string) error {
	oldHashName := Md5Str(md5.Sum([]byte(oldName)))
	newHashName := Md5Str(md5.Sum([]byte(newName)))
	_, err := DB.Exec("UPDATE user SET `Name` = ? WHERE `Name` = ?", newHashName, oldHashName)
	if err != nil {
		return err
	}
	return nil
}

// RePassword 修改密码
func RePassword(DB *sql.DB, Name string, newPassword string) error {
	newHashPassword := Md5Str(md5.Sum([]byte(newPassword)))
	hashName := Md5Str(md5.Sum([]byte(Name)))
	_, err := DB.Exec("UPDATE user SET password = ? WHERE `name` = ?", newHashPassword, hashName)
	if err != nil {
		return err
	}
	return nil
}
