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
	"github.com/Unknwon/goconfig"
	"strconv"
	"bytes"
	"os"
	"bufio"
	"io"
	"strings"
)

var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile
//Mysql Redis初始化
func init() {
	cfg, ConfError = goconfig.LoadConfigFile("config.ini")
	if ConfError != nil {
		panic("配置文件config.ini不存在,请将配置文件复制到运行目录下")
	}
	logLevel, ConfError = cfg.GetValue("Log", "logLevel")
	if ConfError != nil {
		log.SetLevel(log.LevelInfo)
	} else {
		log.SetLevelByName(logLevel)
	}
	username, ConfError = cfg.GetValue("MySQL", "username")
	if ConfError != nil {
		panic("读取数据库username错误")
	}
	password, ConfError = cfg.GetValue("MySQL", "password")
	if ConfError != nil {
		panic("读取数据库password错误")
	}
	url, ConfError = cfg.GetValue("MySQL", "url")
	if ConfError != nil {
		panic("读取数据库url错误")
	}
	address, ConfError = cfg.GetValue("Redis", "address")
	if ConfError != nil {
		panic("读取数据库address错误")
	}
	redis_Pwd, ConfError = cfg.GetValue("Redis", "password")
	if ConfError != nil {
		panic("读取Redis password错误")
	}
	redis_db, ConfError = cfg.GetValue("Redis", "database")
	if ConfError != nil {
		redis_db = "0"
	}
	redis_Database, ConfError = strconv.Atoi(redis_db)
	if ConfError != nil {
		redis_Database = 0
	}
	var dataSourceName bytes.Buffer
	dataSourceName.WriteString(username)
	dataSourceName.WriteString(":")
	dataSourceName.WriteString(password)
	dataSourceName.WriteString("@")
	dataSourceName.WriteString(url)
	db, err = sql.Open("mysql", dataSourceName.String())
	if err != nil {
		log.Error(err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("数据库连接出错,请检查配置账号密码是否正确")
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(30)
	initRedisPool()
	initWriteHasIndexKey();
}

var hasIndexKeys []string
//Redis
var redisPool *redis.Pool

func initRedisPool() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("run time panic: %v", err)
			hasIndexKeys = make([]string, 0)
			file, err := os.OpenFile("hasIndexKeys.txt", os.O_CREATE | os.O_RDONLY, 0666)
			defer file.Close()
			if err == nil {
				reader := bufio.NewReader(file)
				for {
					buf, _, err := reader.ReadLine()
					if err != io.EOF {
						setKeyVal(string(buf), "")
					} else {
						break
					}
				}
				preIndexKeySize = len(hasIndexKeys)
			}

		}
	}()
	redisPool = &redis.Pool{
		MaxIdle:100,
		IdleTimeout: time.Second * 300,
		Dial: func() (redis.Conn, error) {
			var conn redis.Conn
			var cErr error
			if len(redis_Pwd) == 0 {
				conn, cErr = redis.Dial("tcp", address)
				if cErr != nil {
					log.Errorf("Redis初始化失败,请检查配置是否填写正确,key存储切换到文件模式")
					return nil, cErr
				}
			} else {
				conn, cErr = redis.Dial("tcp", address, redis.DialPassword(redis_Pwd), redis.DialDatabase(redis_Database))
				if cErr != nil {
					log.Errorf("Redis初始化失败,请检查配置是否填写正确,key存储切换到文件模式")
					return nil, cErr
				}
			}

			return conn, nil
		},
	}
	DoRedis()
}

const intervalTime = time.Second * 5

var hasIndexKeySize int
var preIndexKeySize int

func initWriteHasIndexKey() {
	if hasIndexKeys != nil {
		go func() {
			ch := time.NewTicker(intervalTime).C
			for {
				<-ch;
				hasIndexKeySize = len(hasIndexKeys)
				tempKeys := hasIndexKeys[preIndexKeySize:hasIndexKeySize]
				preIndexKeySize = hasIndexKeySize
				if len(tempKeys) != 0 {
					file, err := os.OpenFile("hasIndexKeys.txt", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
					if err != nil {
						log.Error(err)
					}
					defer file.Close()
					outputWriter := bufio.NewWriter(file)
					for _, v := range tempKeys {
						outputWriter.WriteString(v + "\n")
					}
					outputWriter.Flush()
				}

			}
		}()
	}
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
	//可以先存几个热门的用户到数据库表avaiuk中 也可以直接GetFollow(2736848922, 0)爬取
	mode, ConfError = cfg.GetValue("Mode", "mode")
	if ConfError != nil {
		panic("读取mode错误")
	} else {
		if m, _ := strconv.Atoi(mode); m == 1 {
			start_uk, err := cfg.GetValue("Mode", "uk")
			if err != nil {
				panic("读取开始爬取uk错误")
			} else {
				log.Info("从单个uk开始爬取")
				s_uk, _ := strconv.ParseInt(start_uk, 10, 64)
				GetFollow(s_uk, 0, true)

			}

		} else {
			log.Info("从数据库存储uk开始爬取")
			for{
				rows, _ := db.Query("select id,flag,uk from avaiuk where flag=0  limit 1")
				if rows.Next() {
					rows.Scan(&id, &flag, &uk)
					stmt, _ := db.Prepare("update avaiuk set flag=1 where id=?")
					stmt.Exec(id)
					log.Info("Select new uk:", uk)
					stmt.Close()
					GetFollow(uk, 0, true)
				}else {
					break
				}
			}

		}
	}
	log.Info("已经递归爬取完成，请切换新的热门uk或者存储新的热门uk到数据库表avaiuk中")
	time.Sleep(time.Second * 2)

}

func checkKeyExist(key interface{}) bool {
	if hasIndexKeys != nil {
		if ok := sliceKeyExist(hasIndexKeys, fmt.Sprintf("%v", key)); ok {
			return true
		} else {
			return false
		}
	} else {
		return RedisKeyExists(key)
	}
}
func sliceKeyExist(s []string, key string) bool {
	for _, v := range s {
		if strings.Compare(v, key) == 0 {
			return true
		}
	}
	return false
}

func setKeyVal(key, val interface{}) {
	if hasIndexKeys != nil {
		hasIndexKeys = append(hasIndexKeys, fmt.Sprintf("%v", key))
	} else {
		RedisSetKV(key, val)
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
	rdsConn := redisPool.Get()
	result, error := rdsConn.Do("ping")
	if error != nil {
		log.Error(error.Error())
		return err.Error()
	}
	return result
}
func RedisSetKV(key interface{}, value interface{}) {
	conn := redisPool.Get()
	defer conn.Close()
	_, error := conn.Do("set", key, value)
	if error != nil {
		log.Error(error.Error())
	}
}
//redis中键是否存在
func RedisKeyExists(key interface{}) bool {
	conn := redisPool.Get()
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
func GetFollow(uk int64, start int, index bool) {
	log.Info("Into uk:", uk, ",start:", start)
	flag := checkKeyExist(uk)
	if (!flag) {
		setKeyVal(uk, "")
		if (index) {
			IndexResource(uk)
		}
		RecursionFollow(uk, start, true)
	} else {
		if start > 0 {
			RecursionFollow(uk, start, false)
		} else {
			log.Warn("Has index UK:", uk)
		}
	}
}

func RecursionFollow(uk int64, start int, goPage bool) {
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
					shareCount := v.Pubshare_count
					if followcount > 0 {
						if (shareCount > 0) {
							GetFollow(v.Follow_uk, 0, true)
						} else {
							GetFollow(v.Follow_uk, 0, false)
						}

					}
				}
				if (goPage) {
					page := (f.Total_count - 1) / 24 + 1
					for i := 1; i < page; i++ {
						GetFollow(uk, 24 * i, false)
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
	Follow_count   int
	Follow_uk      int64
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
	Album_count    int
}

type feedata struct {
	Records []records
}
type records struct {
	Shareid   string
	Title     string
	Feed_type string //专辑：album 文件或者文件夹：share
	Album_id  string
	Category  int
}

var nullstart = time.Now().Unix()
var uinfoId int64 = 0

func IndexResource(uk int64) {
	for true {
		url := "http://pan.baidu.com/wap/share/home?uk=%d&start=%d"
		real_url := fmt.Sprintf(url, uk, 0)

		result, _ := HttpGet(real_url, nil)

		yundata := GetData(result)
		if yundata == nil {
			temp := nullstart
			nullstart = time.Now().Unix()
			if nullstart - temp < 2 {
				log.Warn("被百度限制了 休眠50s")
				time.Sleep(50 * time.Second)
			}
		} else {

			share_count := yundata.Uinfo.Pubshare_count
			album_count := yundata.Uinfo.Album_count
			if share_count > 0 || album_count > 0 {

				res, err := db.Exec("INSERT into uinfo(uk,uname,avatar_url) values(?,?,?)", uk, yundata.Uinfo.Uname, yundata.Uinfo.Avatar_url)
				checkErr(err)
				id, err := res.LastInsertId()

				uinfoId = id
				checkErr(err)
				log.Info("insert uinfo，uk:", uk, ",uinfoId:", uinfoId)

				for _, v := range yundata.Feedata.Records {
					if strings.Compare(v.Feed_type, "share") == 0 {
						db.Exec("insert into sharedata(title,shareid,uinfo_id,category) values(?,?,?,?)", v.Title, v.Shareid, uinfoId, v.Category)
						log.Info("insert share")
					} else if strings.Compare(v.Feed_type, "album") == 0 {
						db.Exec("insert into sharedata(title,album_id,uinfo_id,category) values(?,?,?,?)", v.Title, v.Album_id, uinfoId, v.Category)
						log.Info("insert album")
					}

				}

			}
			totalpage := (share_count + album_count - 1) / 20 + 1
			var index_start = 0
			for i := 1; i < totalpage; i++ {
				index_start = i * 20
				real_url = fmt.Sprintf(url, uk, index_start)
				result, _ := HttpGet(real_url, nil)
				yundata = GetData(result)
				if yundata != nil {
					for _, v := range yundata.Feedata.Records {
						if strings.Compare(v.Feed_type, "share") == 0 {
							db.Exec("insert into sharedata(title,shareid,uinfo_id,category) values(?,?,?,?)", v.Title, v.Shareid, uinfoId, v.Category)
							log.Info("insert share")
						} else if strings.Compare(v.Feed_type, "album") == 0 {
							db.Exec("insert into sharedata(title,album_id,uinfo_id,category) values(?,?,?,?)", v.Title, v.Album_id, uinfoId, v.Category)
							log.Info("insert album")
						}
					}

				} else {
					i--
					temp := nullstart
					nullstart = time.Now().Unix()
					//2次异常小于2s 被百度限制了 休眠50s
					if nullstart - temp < 2 {
						log.Warn("被百度限制了 休眠50s")
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