package common

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net"
	"one-api/lang"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Println(err)
	}
}

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}
	return
}

var sizeKB = 1024
var sizeMB = sizeKB * 1024
var sizeGB = sizeMB * 1024

func Bytes2Size(num int64) string {
	numStr := ""
	unit := "B"
	if num/int64(sizeGB) > 1 {
		numStr = fmt.Sprintf("%.2f", float64(num)/float64(sizeGB))
		unit = "GB"
	} else if num/int64(sizeMB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeMB)))
		unit = "MB"
	} else if num/int64(sizeKB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeKB)))
		unit = "KB"
	} else {
		numStr = fmt.Sprintf("%d", num)
	}
	return numStr + " " + unit
}

func Seconds2Time(num int) (time string) {
	if num/31104000 > 0 {
		time += strconv.Itoa(num/31104000) + lang.T(nil, "utils.time.year")
		num %= 31104000
	}
	if num/2592000 > 0 {
		time += strconv.Itoa(num/2592000) + lang.T(nil, "utils.time.month")
		num %= 2592000
	}
	if num/86400 > 0 {
		time += strconv.Itoa(num/86400) + lang.T(nil, "utils.time.day")
		num %= 86400
	}
	if num/3600 > 0 {
		time += strconv.Itoa(num/3600) + lang.T(nil, "utils.time.hour")
		num %= 3600
	}
	if num/60 > 0 {
		time += strconv.Itoa(num/60) + lang.T(nil, "utils.time.minute")
		num %= 60
	}
	time += strconv.Itoa(num) + lang.T(nil, "utils.time.second")
	return
}

func Interface2String(inter interface{}) string {
	switch inter.(type) {
	case string:
		return inter.(string)
	case int:
		return fmt.Sprintf("%d", inter.(int))
	case float64:
		return fmt.Sprintf("%f", inter.(float64))
	}
	return lang.T(nil, "utils.interface.not_implemented")
}

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}

func IntMax(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func IsIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

func GetUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateRandomCharsKey(length int) (string, error) {
	b := make([]byte, length)
	maxI := big.NewInt(int64(len(keyChars)))

	for i := range b {
		n, err := crand.Int(crand.Reader, maxI)
		if err != nil {
			return "", err
		}
		b[i] = keyChars[n.Int64()]
	}

	return string(b), nil
}

func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length*3/4) // 对于48位的输出，这里应该是36
	if _, err := crand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func GenerateKey() (string, error) {
	//rand.Seed(time.Now().UnixNano())
	return GenerateRandomCharsKey(48)
}

func GetRandomInt(max int) int {
	//rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

func Max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func MessageWithRequestId(message string, id string) string {
	return fmt.Sprintf(lang.T(nil, "utils.error.request_id"), message, id)
}

func RandomSleep() {
	// Sleep for 0-3000 ms
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
}

func GetPointer[T any](v T) *T {
	return &v
}

func Any2Type[T any](data any) (T, error) {
	var zero T
	bytes, err := json.Marshal(data)
	if err != nil {
		return zero, err
	}
	var res T
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return zero, err
	}
	return res, nil
}

// SaveTmpFile saves data to a temporary file. The filename would be apppended with a random string.
func SaveTmpFile(filename string, data io.Reader) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return "", errors.Wrapf(err, lang.T(nil, "utils.error.create_temp_file"), filename)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return "", errors.Wrapf(err, lang.T(nil, "utils.error.copy_temp_file"), filename)
	}

	return f.Name(), nil
}

// GetAudioDuration returns the duration of an audio file in seconds.
func GetAudioDuration(ctx context.Context, filename string) (float64, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 {{input}}
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := cmd.Output()
	if err != nil {
		return 0, errors.Wrap(err, lang.T(nil, "utils.error.get_audio_duration"))
	}

	return strconv.ParseFloat(string(bytes.TrimSpace(output)), 64)
}
