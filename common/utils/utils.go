package utils

import (
	"bygame/common/log"
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/mgo.v2/bson"
)

var jwtKey string

func CreateJwt(id string) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 100).Unix(),
		Id:        id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func CreateJwtZy(id string, uid string, zyUid string) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 100).Unix(),
		Id:        id,
		Subject:   uid,
		Audience:  zyUid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func VerifyJwt(tokenString string) (bool, string) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtKey), nil
	})
	if err != nil {
		return false, ""
	}
	if !token.Valid {
		return false, ""
	}
	if m, ok := token.Claims.(jwt.MapClaims); ok {
		return true, m["jti"].(string)
	}
	return false, ""
}

func init() {
	jwtKey = "lizenghui"
	rand.Seed(time.Now().UnixNano())
}

func RandSeed() int64 {
	return rand.Int63n(int64(8000000000)) + 1000000000
}

func HttpGet(url string) ([]byte, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	res, err := client.Get(url)
	if err == nil {
		defer res.Body.Close()
		return ioutil.ReadAll(res.Body)
	}
	return nil, err
}

func HttpPostJson(url string, body []byte) ([]byte, error) {
	return HttpPostJsonWithTimeOut(url, body, 2)
}

func HttpPostJsonWithTimeOut(url string, body []byte, seconds int) ([]byte, error) {
	client := http.Client{
		Timeout: time.Duration(seconds) * time.Second,
	}
	reader := bytes.NewReader(body)
	res, err := client.Post(url, "application/json; charset=UTF-8", reader)
	if err == nil {
		defer res.Body.Close()
		result, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return result, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numBytes      = "0123456789"
	easyReadBytes = "ABCDEFGHJKLMNPQRSTUVWXYZ123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RandEasyReadCode(size int) string {
	var buf bytes.Buffer
	buf.Grow(size)
	btsLen := len(easyReadBytes)
	for i := 0; i < size; i++ {
		buf.WriteByte(easyReadBytes[RandInt(0, btsLen-1)])
	}
	return buf.String()
}

func RandEasyIntCode(size int) string {
	var buf bytes.Buffer
	buf.Grow(size)
	btsLen := len(numBytes)
	for i := 0; i < size; i++ {
		buf.WriteByte(numBytes[RandInt(0, btsLen-1)])
	}
	return buf.String()
}

func RandInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	return (rand.Int() % (max - min + 1)) + min
}

func RandInt32(min, max int32) int32 {
	if min > max {
		min, max = max, min
	}
	return (rand.Int31() % (max - min + 1)) + min
}

func RandInt64(min, max int64) int64 {
	if min > max {
		min, max = max, min
	}
	return (rand.Int63() % (max - min + 1)) + min
}

/*
生成聊天的会话id
*/
func GetSessionId(uid, uid2 string) string {
	if bson.ObjectIdHex(uid) > bson.ObjectIdHex(uid2) {
		return uid2 + "_" + uid
	}
	return uid + "_" + uid2
}

func Unmarshal(j bool, data []byte, m protoreflect.ProtoMessage) error {
	if j {
		return json.Unmarshal(data, m)
	} else {
		return proto.Unmarshal(data, m)
	}
}

func Marshal(j bool, m protoreflect.ProtoMessage) ([]byte, error) {
	if j {
		return json.Marshal(m)
	} else {
		return proto.Marshal(m)
	}
}

func GetCardValue(card int64) (int64, int64) {
	return card / 16, card % 16
}

func FindIndex[T int | int64 | int32 | string](slc []T, target T) (index int, ok bool) {
	for k, v := range slc {
		if v == target {
			return k, true
		}
	}
	return
}

func InsertIndex[T int | int64 | int32 | string](slc []T, index int, target T) (slc2 []T) {
	if index > len(slc)-1 || index < 0 {
		index = 0
	}
	slc2 = append(slc2, slc[:index]...)
	slc2 = append(slc2, target)
	slc2 = append(slc2, slc[index:]...)
	return
}

// 万分之概率
func Rate1w(target int) bool {
	return RandInt(1, 10000) <= target
}

// 一分之概率
func Rate1(target float32) bool {
	if target >= 1 {
		return true
	}
	if target <= 0 {
		return false
	}
	i := 10000 * 10000 * 10000
	return RandInt(1, i) <= int(target*float32(i))
}

func Reqzy(router string, sign string, query map[string]string) ([]byte, error) {
	// sign = "&p=android&v=27&c=dev&vs=1.1.27.0922&_uid=545&_mtime=1697615049149&_random=HFdOqx+l0vdWf+9WSMQ7hwaaqKSvCSuf\n&_public=ey0JSi0+KHXSb2sA6il8rtSXdPLZKEHhZ34PlWG3YzA=\n&_req_id=545-67e4dd5d-c57b-49f0-8c02-641ed0d69e6b&_sign=5NDdMgWINNfSckGMHSS5tc8v+BaOIY/B7h6sVTMnQ5YcCCeQPk37lKgJVMBJuUm59Y4jIsuyODdr\nlXNYAfIlB0Yw2+BKvaTV3v6BgnxlN4Yov7RvoRYO3zpTHnz+qH1jLwQCDGmPvsvtPCQzo5EKetL/\n82vFfoEB5rvqYtq3Iz7TWK22bQyk+73RqS7PEN4Q3sCC2AYOs8hf2JbhQz5vzsBg3oA0Cs6P1KNk\nYpcuINIvkO2aXLwJDaxS9EZq0s5yerl5+Q/h6X9EYTLJZselL6Z2TptkmxlByYaZ+u204pCLAdkt\nnFrOUDCh8M/HqCANkMivO915Wrz2DLxSEPbRBb4nA0iV2YNcbxi4BeZp6YOuFx9wrslsIQmQ8GkH\nCWHl0qETAwARyy4lHVxIeRBTofTxe81HlX8riPnNvIAHsfjKoJqkH/4xDg1x72zstfDPNWiexJHy\nKpSitKkGMLHgwemiGh/S93rfDHcrCbq0Q6ZmjNw5qVNehc5/eNfqL3B85RQebPozDoNMduaMqB74\nXLvki6XJCKRR+C34BA==\n&smei_id=202310101637468e2a5db2ed8d13de5a1f01b18f37f72601120ad4ab88443e&lang=zh-Hans-CN&user_lang="
	prikey := "pB1AnMp2IkYRwftUgT0yA5JtsG1tUw6Bchh5Wg0yCFY="
	slc := strings.Split(sign, "&")
	keys := make(map[string]string)
	for _, data := range slc {
		index := strings.Index(data, "=")
		if index != -1 && index < len(data) {
			keys[data[:index]] = data[index+1:]
		}
	}

	keyMap := map[string]string{
		"v":           keys["v"],
		"c":           keys["c"],
		"p":           keys["p"],
		"_mtime":      keys["_mtime"],
		"_public":     keys["_public"],
		"_random":     keys["_random"],
		"_uid":        keys["_uid"],
		"_sign":       keys["_sign"],
		"game_v":      "1",                                // 游戏版本
		"game_id":     "9",                                // 游戏ID
		"open_id":     "1",                                // provider的id
		"open_mtime":  fmt.Sprint(time.Now().UnixMilli()), // 签名时间
		"open_req_id": bson.NewObjectId().Hex(),           // 请求ID（不可重复请求）
		"private_key": prikey,                             // 服务商私钥
	}

	hashStr := fmt.Sprintf(
		"%v%v%v%v%v%v%v%v%v%v%v",
		prikey,
		keyMap["open_id"],
		keyMap["open_mtime"],
		keyMap["open_req_id"],
		keyMap["game_v"],
		keyMap["game_id"],
		keyMap["_mtime"],
		keyMap["_public"],
		keyMap["_random"],
		keyMap["_sign"],
		keyMap["_uid"],
	)

	hasher := sha512.New()
	hasher.Write([]byte(hashStr))
	hash := hasher.Sum(nil)
	keyMap["open_sign"] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", hash)))

	for k, v := range query {
		keyMap[k] = v
	}

	urlStr := "https://z.miyu520.com" + router + "?"

	for k, v := range keyMap {
		if v != "" {
			urlStr += fmt.Sprintf("%v=%v&", k, url.QueryEscape(v))
		}
	}
	urlStr = urlStr[:len(urlStr)-1]
	return HttpGet(urlStr)
}

type RandomItem struct {
	Id     int32
	Weight float32
}

// 加权随机
func RandomWeight(items []RandomItem) int32 {
	// 计算总权重
	var totalWeight float32
	for _, item := range items {
		totalWeight += item.Weight
	}
	result := float64(RandInt(1, int(totalWeight*1000*1000*1000))) / float64(1000*1000*1000)

	var tmpWeight float32
	for _, item := range items {
		tmpWeight += item.Weight
		if tmpWeight >= float32(result) {
			return item.Id
		}
	}
	return 0
}

func Max[T int | int64 | int32 | string](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

func Min[T int | int64 | int32 | string](a, b T) T {
	if a <= b {
		return a
	}
	return b
}

/*
	pop

	得到access token用于后续接口请求
	http(s)://<ip>:<port>/user/token


*/

// pop 钱包对接
func PopTest() {
	// addr := "http://23.105.205.207:5003/user/token"
	// popUserToken()
	// popUserBalance()
	// popAddCoin()
	// popDeduct()
}

// func popUserToken() {
// 	addr := "http://23.105.205.207:5003/user/token"
// 	var req reqPopUserToken
// 	req.AppKey = "g7ejjPtz"
// 	req.Time = fmt.Sprint(time.Now().Unix())
// 	req.IdToken = "3dd84c2f9346600dc1964eef01d35e3f2d080bbc5e656336db0400db962c68ee"
// 	sign := PopSign(req)
// 	req.Sign = sign
// 	bts, err := json.Marshal(req)
// 	b, err2 := HttpPostJson(addr, bts)
// 	var ret retPopUserToken
// 	json.Unmarshal(b, &ret)
// 	fmt.Printf("req: %+v,ret: %+v,err: %v,err2: %v msg:%v,bts:%v\n", req, ret, err, err2, ret.Message, string(b))
// 	// fmt.Println(ret.Data.AccessToken)
// 	fmt.Printf("获取用户基础信息:\nhttp://23.105.205.207:5003/user/balance?access_token=%v\n", ret.Data.AccessToken)
// }

type reqPopUserToken struct {
	AppKey  string `json:"k"`
	Time    string `json:"t"` // s
	IdToken string `json:"id"`
	Sign    string `json:"sign"`
}

type retPopUserToken struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    retPopUserTokenData `json:"data"`
}

type retPopUserTokenData struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

// 计算pop的签名
func PopSign(req any, appSecret string) (signStr string) {
	if req == nil {
		return
	}
	b, err := json.Marshal(req)
	if err != nil {
		log.Wrn("pop 签名计算错误 req: %v", req)
		return
	}
	var m map[string]any
	err2 := json.Unmarshal(b, &m)
	if err2 != nil {
		log.Wrn("pop 签名计算错误 req: %v", req)
		return
	}

	// 正序拼接
	keys := make([]string, 0, len(m))
	for k := range m {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	str := ""

	for _, mk := range keys {
		if str != "" {
			str += "&"
		}
		str = fmt.Sprintf("%s%s=%v", str, mk, fmt.Sprint(m[mk]))
	}

	fmt.Println(str)

	hasher := sha512.New()
	hasher.Write([]byte(str + appSecret))
	hash := hasher.Sum(nil)
	signStr = fmt.Sprintf("%x", hash)
	return
}

// user/balance?access_token=ACCESS_TOKEN

func popUserBalance() {
	// addr := fmt.Sprintf("http://23.105.205.207:5003/user/balance?access_token=%v", )
	// fmt.Println(addr)
	// bts, err := HttpGet(addr)
	// fmt.Println(string(bts), err)
}

func GetArgs(key string) string {
	if len(os.Args) > 1 {
		for _, kv := range os.Args {
			index := strings.Index(kv, "=")
			if index != -1 && kv[:index] == key {
				return kv[index+1:]
			}
		}
	}
	return ""
}
