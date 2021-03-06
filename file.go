package gotool

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

//删除目录
func Remove(path string) bool {
	err := os.RemoveAll(path)
	if err == nil {
		return true
	} else {
		return false
	}
}

//判断文件是否存在
func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

//写入内容到文件(存在则追加)
func WriteByteToFile(data []byte, file string) (size int, err error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	size, err = w.Write(data)
	w.Flush()
	return size, err
}

//计算文件Md5
func FileMd5(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b := bufio.NewReader(f)

	m := md5.New()
	_, err = io.Copy(m, b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", m.Sum(nil)), err

}

//复制单文件(bufio缓存)
func CopyFile(srcFile string, destFile string) (int64, error) {
	sf, err := os.Open(srcFile)
	if err != nil {
		return 0, err
	}
	defer sf.Close()

	buf := bufio.NewReader(sf)
	df, err := os.Create(destFile)
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(df)
}

//遍历目录获取文件列表
type Dir struct {
	Name  string
	Files []os.FileInfo
}

var dirInfo []Dir

func RangeDir(dir string) ([]Dir, error) {
	var files []os.FileInfo
	filesInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return dirInfo, err
	}

	for _, f := range filesInfo {
		if f.IsDir() {
			_, err = RangeDir(dir + "/" + f.Name())

		} else {
			files = append(files, f)
		}

	}
	newDir := Dir{
		Name:  dir,
		Files: files,
	}
	dirInfo = append(dirInfo, newDir)
	return dirInfo, err
}

//递归复制目录
func CopyDir(srcDir string, destDir string) (err error) {
	if IsExist(destDir) == false {
		os.Mkdir(destDir, 0755)
	}
	d, err := ioutil.ReadDir(srcDir)
	for _, file := range d {
		if file.IsDir() {
			path := srcDir + "/" + file.Name()
			go CopyDir(path, destDir+"/"+file.Name())
		} else {
			_, err = CopyFile(srcDir+"/"+file.Name(), destDir+"/"+file.Name())
		}
	}
	return err
}

//逐行读取文本
func ReadLine(fileName string, handler func(string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

//本地图片转base64
func PicToBase64(src string) (string, error) {
	imgFile, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(imgFile), nil
}

//网络图片转base64(5秒超时)
func PicToBase64ByUrl(url string) (string, error) {
	if url == "" {
		return "", errors.New("下载链接为空")
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	avatarResp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	avatarBody, err := ioutil.ReadAll(avatarResp.Body)
	if err != nil {
		return "", err
	}
	defer avatarResp.Body.Close()

	return base64.StdEncoding.EncodeToString(avatarBody), nil
}
