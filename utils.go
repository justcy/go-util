package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

func CheckErr(err error) {
	if err != nil {
		fmt.Println("err :", err)
	}
}
func UTCTime() int64 {
	return time.Now().UTC().Unix()
}

//获取当前时间戳
func GetNowUnix() int64 {
	return time.Now().Unix()
}

//获取当前时间，单位纳秒
func GetNowUnixNano() int64 {
	return time.Now().UnixNano()
}

//MD5 计算字符串MD5值
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
