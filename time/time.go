package main

import (
	"fmt"
	"time"
)

func main()  {
	TimeUsage()
	TimeOperate()
	Ticker()
}


// 1.获取基础属性
func TimeUsage()  {
	now := time.Now()
	fmt.Println(
		now.Year(),		// 2021
		now.Month(),	// September
		now.Day(),		// 14
		now.Hour(),		// 22
		now.Minute(),	// 30
		now.Second(),	// 35
		now.Unix(),		// 获取1970-01-01:00:00:00 到现在经过的秒数(时间戳)
		now.UnixNano(),	// 获取纳秒
	)

	fmt.Println(now.Date())	// 获取日期:2021 September 14
}

// 2.常见操作
func TimeOperate()  {
	start := time.Now()
	time.Sleep(5*time.Second)

	// a.计算时间差
	fmt.Println(time.Now().Sub(start))	// 5.0004553s

	// b.计算时间和
	fmt.Println(start.Add(-1 * time.Hour))					// 一小时前	2021-09-14 21:44:14.4853463 +0800 CST m=-3599.992866799
	fmt.Println(start.AddDate(0,0,-1))	// 一天以前	2021-09-13 22:44:14.4853463 +0800 CST

	// c.将Time转换化为字符串
	fmt.Println(start.Format(time.RFC822))					// 14 Sep 21 22:44 CST
	fmt.Println(start.Format("2006-01-02 15:04:05"))	// 2021-09-14 22:44:14

	// d.将字符串转换为Time
	// 注意:layout的样式一定要和要解析的时间字符串一致,否则会解析失败

	timeObj1,err :=time.Parse("2006/01/02 15:04:05","2021/09/12 14:32:21")	// 使用UTC时间进行解析
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("timeObj1:",timeObj1)	// timeObj1: 2021-09-12 14:32:21 +0000 UTC

	// 加载时区
	loc,err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return
	}
	timeObj2,err := time.ParseInLocation("2006-01-02 15:04:05","2021-08-06 13:12:44",loc)
	if err != nil {
		return
	}
	fmt.Println("timeObj2:",timeObj2)	// timeObj2: 2021-08-06 13:12:44 +0800 CST


}

// 3.定时器
func Ticker()  {
	ticker := time.NewTicker(time.Second) //	定义一个1秒间隔的定时器
	defer ticker.Stop()					  //	关闭定时器
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()

	for {
		select {
		case <- done:
			fmt.Println("Done!")
			return
		case t := <- ticker.C:
			fmt.Println("Current time:",t)
		}
	}
	/*
	Current time: 2021-09-14 23:21:06.1068534 +0800 CST m=+6.073710801
	Current time: 2021-09-14 23:21:07.1063597 +0800 CST m=+7.073217001
	Current time: 2021-09-14 23:21:08.1069143 +0800 CST m=+8.073771701
	Current time: 2021-09-14 23:21:09.1064671 +0800 CST m=+9.073324501
	Current time: 2021-09-14 23:21:10.1078919 +0800 CST m=+10.074749301
	Current time: 2021-09-14 23:21:11.1065265 +0800 CST m=+11.073383801
	Current time: 2021-09-14 23:21:12.1079344 +0800 CST m=+12.074791801
	Current time: 2021-09-14 23:21:13.107165 +0800 CST m=+13.074022401
	Current time: 2021-09-14 23:21:14.1071793 +0800 CST m=+14.074036701
	Current time: 2021-09-14 23:21:15.1139324 +0800 CST m=+15.080789801
	Done!
	...
	*/
}