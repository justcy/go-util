package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func RandGroup(p []uint32) int {
	if p == nil {
		panic("args not found")
	}
	rl := p[len(p)-1]
	if rl == 0 {
		return 0
	}
	rn := uint32(rand.Int63n(int64(rl)))
	for i := 0; i < len(p); i++ {
		if rn < p[i] {
			return i
		}
	}
	panic("bug")
}
func RandGroupN(p ...uint32) int {
	if p == nil {
		panic("args not found")
	}
	r := make([]uint32, len(p))
	for i := 0; i < len(p); i++ {
		if i == 0 {
			r[0] = p[0]
		} else {
			r[i] = r[i-1] + p[i]
		}
	}
	return RandGroup(r)
}

//随机获取map中的一个元素，map的key为uint32代表权重
func RandGroupItem(m map[uint32]interface{}) interface{} {
	if m == nil {
		panic("args not found")
	}
	r := make([]uint32, 0, len(m))
	for key := range m {
		r = append(r, key)
	}
	return m[uint32(RandGroup(r))]
}

//在a-b之间随机获取一个整数
func RandInterval(a, b int32) int32 {
	if a == b {
		return a
	}
	min, max := int64(a), int64(b)
	if min > max {
		min, max = max, min
	}
	return int32(rand.Int63n(max-min+1) + min)
}

//随机获取0-max之间的一个随机数
func RandNum(max int32) int32 {
	return RandInterval(int32(0), int32(max))
}

//在a,b之间随机获取最多num个整数
func RandIntervalN(a, b int32, num uint32) []int32 {
	if a == b {
		return []int32{a}
	}
	min, max := int64(a), int64(b)
	if min > max {
		min, max = max, min
	}
	l := max - min + 1
	if int64(num) > l {
		num = uint32(l)
	}
	r := make([]int32, num)
	m := make(map[int32]int32)
	for i := uint32(0); i < num; i++ {
		v := int32(rand.Int63n(l) + min)

		if mv, ok := m[v]; ok {
			r[i] = mv
		} else {
			r[i] = v
		}

		lv := int32(l - 1 + min)
		if v != lv {
			if mv, ok := m[lv]; ok {
				m[v] = mv
			} else {
				m[v] = lv
			}
		}

		l--
	}
	return r
}
//随机打乱顺序
func Shuffle(slice []int32 ) []int32 {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i],slice[j] = slice[j],slice[i]
	})
	return slice
}
func HitRate100(rate int32) bool {
	return RandNum(100) < rate
}
func HitRate1000(rate int32) bool {
	return RandNum(1000) < rate
}
func HitRate10000(rate int32) bool {
	return RandNum(10000) < rate
}


func randString(length int, source []byte) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = source[r.Intn(len(source))]
	}
	return string(bytes)
}
func RandStringNumber(length int) string {
	var temp = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	return randString(length, temp)
}
func RandString(length int) string {
	var temp = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	return randString(length, temp)
}
