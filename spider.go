package main

import (
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/siddontang/go/log"
	"regexp"
	"encoding/json"
	"time"
	"database/sql"
	"github.com/garyburd/redigo/redis"
)

var db *sql.DB
var err error
//Mysql初始化
func init() {
	db, err = sql.Open("mysql", "root:123456@(127.0.0.1:3306)/baidu")
	if err!=nil{
		log.Error("数据库连接出错")
	}
	db.SetMaxOpenConns(50)
}
//Redis
func GetRedisPool() *redis.Pool {
	rdspool := &redis.Pool{
		MaxIdle:100,
		IdleTimeout: time.Second * 300,
		Dial: func() (redis.Conn, error) {
			conn, cErr := redis.Dial("tcp", "127.0.0.1:6379", redis.DialPassword("123456"))
			if cErr != nil {
				return nil, cErr
			}
			return conn, nil
		}, }
	return rdspool
}
type sharedata struct {
	Id      int64
	Title   string
	UinfoId int64
	Shareid string
}

func main() {
	var id int64
	var flag int
	var uk int64
	//GetFollow(2736848922, 0)
	//fmt.Println(DoRedis())
	//可以先存几个热门的用户到数据库表avaiuk中 也可以直接GetFollow(2736848922, 0)爬取
	for {
		rows, _ := db.Query("select id,flag,uk from avaiuk where flag=0  limit 1")
		for rows.Next() {
			rows.Scan(&id, &flag, &uk)
		}
		stmt, _ := db.Prepare("update avaiuk set flag=1 where id=?")
		stmt.Exec(id)
		log.Warn("Select new uk:", uk)
		stmt.Close()
		GetFollow(uk, 0,true)
	}

}

func record(rows *sql.Rows) map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		//将行数据保存到record字典
		err = rows.Scan(scanArgs...)
		record := make(map[string]interface{})
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}
		fmt.Println(record)
		return record
	}
	return nil
}

func DoRedis() interface{} {
	pool := GetRedisPool()
	defer pool.Close()
	rdsConn := pool.Get()
	result, error := rdsConn.Do("ping")
	if error != nil {
		log.Error(error.Error())
		return nil
	}
	return result
}
func SetKV(key interface{}, value interface{}) {
	pool := GetRedisPool()
	defer pool.Close()
	conn := pool.Get()
	defer conn.Close()
	_, error := conn.Do("set", key, value)
	if error != nil {
		log.Error(error.Error())
	}
}
//redis中键是否存在
func KeyExists(key interface{}) bool {
	pool := GetRedisPool()
	defer pool.Close()
	conn := pool.Get()
	defer conn.Close()
	result, error := conn.Do("exists", key)
	if error != nil {
		log.Error(error.Error())
		return true
	}
	if result == int64(1) {
		return true
	}
	return false
}


//获取订阅用户
func GetFollow(uk int64, start int,index bool) {
	log.Info("Into uk:",uk,",start:",start)
	flag := KeyExists(uk)
	if (!flag) {
		SetKV(uk, "")
		if(index){
			IndexResource(uk)
		}
		recFollow(uk, start,true)
	} else {
		if start > 0 {
			recFollow(uk, start,false)
		} else {
			log.Warn("Has index UK:", uk)
		}
	}
}

func recFollow(uk int64, start int,goPage bool) {
	url := "http://yun.baidu.com/pcloud/friend/getfollowlist?query_uk=%d&limit=24&start=%d&bdstoken=e6f1efec456b92778e70c55ba5d81c3d&channel=chunlei&clienttype=0&web=1&logid=MTQ3NDA3NDg5NzU4NDAuMzQxNDQyMDY2MjA5NDA4NjU=";
	time.Sleep(time.Second * 5)
	real_url := fmt.Sprintf(url, uk, start)
	result, error := HttpGet(real_url, headers)
	if error == nil {
		var f follow
		error := json.Unmarshal([]byte(result), &f)
		if error == nil {
			if f.Errno == 0 {
				for _, v := range f.Follow_list {
					followcount := v.Follow_count
					shareCount:=v.Pubshare_count
					if followcount > 0 {
						if(shareCount>0){
							GetFollow(v.Follow_uk, 0,true)
						}else {
							GetFollow(v.Follow_uk, 0,false)
						}

					}
				}
				if(goPage){
					page := (f.Total_count - 1) / 24 + 1
					for i := 1; i < page; i++ {
						GetFollow(uk, 24 * i,false)
					}
				}

			} else {
				//被百度限制了 休眠50s
				time.Sleep(time.Second * 50)
			}
		}
	}
}

type follow struct {
	//Request_id int64
	Total_count int
	Follow_list []follow_list
	Errno       int
}
type follow_list struct {
	Pubshare_count int
	Follow_count int
	Follow_uk    int64
}

var headers = map[string]string{
	"User-Agent":"MQQBrowser/26 Mozilla/5.0 (Linux; U; Android 2.3.7; zh-cn; MB200 Build/GRJ22; CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"Referer":"https://yun.baidu.com/share/home?uk=325913312#category/type=0"}


func HttpGet(url string, headers map[string]string) (result string, err error) {

	client := &http.Client{}
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("数据读取异常")
		return "", err
	}
	defer resp.Body.Close()
	return string(body), nil
}

type yundata struct {
	Feedata feedata
	Uinfo   uinfo
}
type uinfo struct {
	Uname          interface{}
	Avatar_url     string
	Pubshare_count int
}

type feedata struct {
	Records []records
}
type records struct {
	Shareid string
	Title   string
}

var nullstart = time.Now().Unix()
var uinfoId int64 = 0

func IndexResource(uk int64) {
	for true {
		url := "http://pan.baidu.com/wap/share/home?uk=%d&start=%d&adapt=pc&fr=ftw"
		real_url := fmt.Sprintf(url, uk, 0)

		result, _ := HttpGet(real_url, nil)

		yundata := GetData(result)
		if yundata == nil {
			temp := nullstart
			nullstart = time.Now().Unix()
			if nullstart - temp < 2 {
				time.Sleep(50 * time.Second)
			}
		} else {

			share_count := yundata.Uinfo.Pubshare_count
			if share_count > 0 {

				res, err := db.Exec("INSERT into uinfo(uk,uname,avatar_url) values(?,?,?)", uk, yundata.Uinfo.Uname, yundata.Uinfo.Avatar_url)
				checkErr(err)
				id, err := res.LastInsertId()

				uinfoId = id
				checkErr(err)
				log.Info("insert uinfo，uk:", uk, ",uinfoId:", uinfoId)

				for _, v := range yundata.Feedata.Records {
					//stmt, _ =db.Prepare("")
					res, _ = db.Exec("insert into sharedata(title,shareid,uinfo_id) values(?,?,?)", v.Title, v.Shareid, id)
					log.Info("insert sharedata")
				}

			}
			totalpage := (share_count - 1) / 20 + 1
			var index_start = 0
			for i := 1; i < totalpage; i++ {
				index_start = i * 20
				real_url = fmt.Sprintf(url, uk, index_start)
				result, _ := HttpGet(real_url, nil)
				yundata = GetData(result)
				if yundata != nil {
					for _, v := range yundata.Feedata.Records {
						db.Exec("insert into sharedata(title,shareid,uinfo_id) values(?,?,?)", v.Title, v.Shareid, uinfoId)
						log.Info("insert sharedata")
					}

				} else {
					i--
					temp := nullstart
					nullstart = time.Now().Unix()
					//2次异常小于2s 被百度限制了 休眠50s
					if nullstart - temp < 2 {
						time.Sleep(50 * time.Second)
					}
				}

			}
			break
		}

	}
}

func GetData(res string) *yundata {
	r, _ := regexp.Compile("window.yunData = (.*})")
	match := r.FindStringSubmatch(res)
	if len(match) < 1 {
		return nil
	}
	var yd yundata
	error := json.Unmarshal([]byte(match[1]), &yd)
	if error != nil {
		return nil
	}
	return &yd
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}