package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func SafeCreateDir(path string) {
	if exists := PathExists(path); exists {
		return
	}
	os.Mkdir(path, os.ModePerm)
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// 读取文件
func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

type FileList struct {
	FilePath string
	FileDir  string
}

func prepareDirs(dirs []string) ([]string, [][]os.FileInfo) {
	resultDir := make([]string, 0)
	resultDirFileInfo := make([][]os.FileInfo, 0)
	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				continue
			}
			if err = os.MkdirAll(dir, 0700); err != nil {
				continue
			}
		} else if !fi.IsDir() {
			continue
		}
		if fis, err := ioutil.ReadDir(dir); err != nil {
		} else {
			resultDir = append(resultDir, dir)
			resultDirFileInfo = append(resultDirFileInfo, fis)
		}
	}
	return resultDir, resultDirFileInfo
}

func GetFileList(fileDir, suffix string) []*FileList {
	//isRecursion := true
	//if len(fileDir) != 0 && fileDir[len(fileDir)-1] == '*' {
	//	isRecursion = true
	//	fileDir = fileDir[:len(fileDir)-2]
	//}
	arrFileLists := make([]*FileList, 0)
	suffixUp := strings.ToUpper(suffix)
	arrDirs, arrInfos := prepareDirs([]string{fileDir})
	for idx, dbDir := range arrDirs {
		for _, fi := range arrInfos[idx] {
			fileName := fi.Name()
			if fi.IsDir() {
				fileList := GetFileList(path.Join(dbDir, fileName), suffix)
				if len(fileList) != 0 {
					arrFileLists = append(arrFileLists, fileList...)
				}
				continue
			}
			// try match suffix and `ordinal_pubKey_bitLength.suffix`
			if !strings.HasSuffix(strings.ToUpper(fileName), suffixUp) {
				continue
			}
			filePath := filepath.Join(dbDir, fileName)
			arrFileLists = append(arrFileLists, &FileList{
				FilePath: filePath,
				FileDir:  dbDir,
			})
		}
	}
	return arrFileLists
}

// CopyFile 复制文件
func CopyFile(orgFile, newFile string) error {
	orgF, err := os.Open(orgFile)
	if err != nil {
		return err
	}
	defer func() { _ = orgF.Close() }()
	newF, err := os.Create(newFile)
	if err != nil {
		return err
	}
	defer func() { _ = newF.Close() }()
	_, err = io.Copy(newF, orgF)
	if err != nil {
		return err
	}
	return newF.Sync()
}

// MoveFile 移动文件
func MoveFile(orgFile, newFile string) error {
	err := CopyFile(orgFile, newFile)
	if err != nil {
		return err
	}
	return os.Remove(orgFile)
}

// GetFileSize file size
func GetFileSize(fileName string) int64 {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err == nil {
		defer func() {
			_ = file.Close()
		}()
		fi, err := file.Stat()
		if err != nil {
			return 0
		}

		return fi.Size()
	}
	return 0
}
