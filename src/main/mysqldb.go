package main
import (
	"arpscan/src/conf"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"
	"strconv"
	"time"
)

var db *gorm.DB

type MacList struct {
	ID        int    `gorm:"primary_key"`
	Ip      string `gorm:"type:varchar(128)"`
	Mac     string `gorm:"type:varchar(128)"`
	CreatedAt time.Time
}

type Ipallcator struct {
	Id int `gorm:"primary_key"`
	Host_name string `gorm:"type:varchar(64)"`
	Pod_name string `gorm:"type:varchar(64)"`
	Container_id string `gorm:"type:varchar(164)"`
	Namespaces string `gorm:"type:varchar(128)"`
	Ip string `gorm:"type:varchar(64)"`
	Status int `gorm:"type:tinyint(1)"`
	Created string `gorm:"type:varchar(64)"`
}

type IpallcatorPage struct {
	Podinfoes []Ipallcator
	Count int
}


func ConnMsql() {
	var err error
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", conf.GetInst().User, conf.GetInst().Password, conf.GetInst().Address, conf.GetInst().Port, conf.GetInst().DBName)
	db, err = gorm.Open("mysql", connArgs)
	if err != nil{
		log.Fatal(err)
	}
}


func CreateTab(){
	Db := db
	if !Db.HasTable(&MacList{}) {
		if err := Db.CreateTable(&MacList{}).Error; err != nil {
			panic(err)
		}
	}

}

func  QueryIpMac(ip string,mac string)(bool){
	Db := db
	var maclist MacList
	Db.Where("ip = ? AND mac = ?", ip, mac).Find(&maclist)
	if (MacList{}) != maclist{
		return true
	}else{
		return false
	}
}

func  QueryIp(ip string)(iplist []MacList){
	Db := db
	Db.Where("ip = ?", ip).Find(&iplist)
	return  iplist
}

func InsertDb(ip string,mac string){
	var dbins = MacList{Ip:ip,Mac:mac,CreatedAt:time.Now()}
	db.Create(&dbins)
}

func QuertPodName(namespace, podname, ip, status string, page,pageSize int) []Ipallcator{
	podinfoes := make([]Ipallcator, 0)
	Db := db

	if namespace != "" {
		Db = Db.Where("namespaces = ?", namespace)
	}

	if podname != "" {
		Db = Db.Where("pod_name = ?", podname)
	}

	if ip != "" {
		Db = Db.Where("ip = ?", ip)
	}

	if status != "" {
		statusNum, _ := strconv.Atoi(status)
		Db = Db.Where("status = ?", statusNum)
	}

	if page > 0 && pageSize > 0 {
		Db = Db.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	if err := Db.Find(&podinfoes).Error; err != nil {
		fmt.Println(err.Error())
	}
	//db.Where("namespaces = ? AND pod_name = ?",namespace,podname).Find(&podinfo)
	return podinfoes
}

func CountPodName(namespace, podname, ip, status string) int{
	podinfoes := make([]Ipallcator, 0)
	Db := db
	count := 0

	if namespace != "" {
		Db = Db.Where("namespaces = ?", namespace)
	}

	if podname != "" {
		Db = Db.Where("pod_name = ?", podname)
	}

	if ip != "" {
		Db = Db.Where("ip = ?", ip)
	}

	if status != "" {
		statusNum, _ := strconv.Atoi(status)
		Db = Db.Where("status = ?", statusNum)
	}

	Db.Find(&podinfoes).Count(&count)
	return count
}