package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func httpGet(url string) string {
	request, err := http.Get(url)
	if err != nil {
		fmt.Println("http get error.")
	}
	defer request.Body.Close()
	html, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("http read error")
	}
	return string(html)
}

func getPV() {

	//获取源码
	htmlPV := httpGet("https://c.163.com/wiki/index.php?title=特殊:热点页面&limit=200&offset=0")

	//正则处理
	regPV := regexp.MustCompile(`title=".*">(.*)<\/a>‏‎（(.*)次浏览）<\/li>`)
	listPV := regPV.FindAllStringSubmatch(htmlPV, -1)
	//fmt.Printf("%q", listPV[0][1])

	//获取已有列名
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/wiki?charset=utf8")
	rows, err := db.Query("select  column_name from information_schema.columns where  table_name = 'pv_monitor';")
	if err != nil {
	}
	listColumnName := map[string]string{}
	for rows.Next() {
		var titlename string
		rows.Columns()
		err = rows.Scan(&titlename)
		listColumnName[titlename] = "0"
	}
	delete(listColumnName, "get_time") //排除get_time
	delete(listColumnName, "id")

	//字符串处理
	var sqlAlter string
	var sqlColumn string
	var sqlValue string
	var IsNew = false
	//判断新增文档
	for m := 0; m < len(listPV); m++ {
		newTitle := listPV[m][1]
		_, exists := listColumnName[newTitle]
		if exists {
			listColumnName[newTitle] = strings.Replace(listPV[m][2], ",", "", -1)
		} else {
			IsNew = true
			sqlAlter += ",`" + newTitle + "` int"
			listColumnName[newTitle] = strings.Replace(listPV[m][2], ",", "", -1)
		}

	}
	//增加列
	if IsNew {
		exec, err := db.Exec("ALTER TABLE pv_monitor ADD(" + strings.TrimPrefix(sqlAlter, ",") + ")")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(exec)
	}
	//遍历
	for key, value := range listColumnName {
		sqlColumn += ",`" + key + "`"
		sqlValue += "," + value
	}
	//插入列
	sql := "insert into pv_monitor(get_time" + sqlColumn + ") values(null" + sqlValue + ")"
	fmt.Println(sql)
	query, err := db.Query(sql)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(query)

	//统计部分
	htmlStc := httpGet("https://c.163.com/wiki/index.php?title=特殊:统计信息")
	htmlStc = strings.Replace(htmlStc, ",", "", -1)
	//正则
	listStc := regexp.MustCompile(`<td class="mw-statistics-numbers">([\d,.]*)<`).FindAllStringSubmatch(htmlStc, -1)
	sqlStc := fmt.Sprintf("insert into statistic_monitor values(null,null,%s,%s,%s,%s,%s,%s,%s)",
		listStc[0][1], listStc[1][1], listStc[2][1], listStc[3][1], listStc[4][1], listStc[10][1], listStc[11][1])
	fmt.Println(sqlStc)
	//入库
	queryStc, err := db.Query(sqlStc)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(queryStc)
}

func main() {
	c := time.Tick(time.Minute)
	for now := range c {
		fmt.Println(now.Minute())
		if now.Minute() == 00 {
			getPV()
		}
	}

}
