/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func IsSymLink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

func CountDirFiles(dirName string) int {
	if !IsDir(dirName) {
		return 0
	}
	var count int
	err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		logger.Fatalf("count dir files failed %v", err)
		return 0
	}
	return count
}

func FileMD5(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	m := md5.New()
	if _, err := io.Copy(m, file); err != nil {
		return "", err
	}

	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fileMd5, nil
}

func DeleteFile(path string) error {
	return os.Remove(path)
}

func LocalMd5Sum(src string) string {
	md5Str, err := FileMD5(src)
	if err != nil {
		logger.Fatalf("get file md5 failed %v", err)
		return ""
	}
	return md5Str
}

func MkFileFullPathDir(fileName string) error {
	localDir := filepath.Dir(fileName)
	err := Mkdir(localDir)
	if err != nil {
		return fmt.Errorf("create local dir %s failed: %v", localDir, err)
	}
	return nil
}

func Mkdir(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

func WriteFile(fileName string, content []byte) error {
	dir := filepath.Dir(fileName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(fileName, content, 0644); err != nil {
		return err
	}
	return nil
}

func AppendFile(fileName string, content string) error {
	if !IsExist(fileName) {
		return errors.Errorf("file %s not exist", fileName)
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return err
	}

	return nil
}

func Tar(src, dst, trimPrefix string) error {
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path, trimPrefix)
		fmt.Println(strings.TrimPrefix(path, string(filepath.Separator)))

		hdr.Name = strings.TrimPrefix(path, string(filepath.Separator))
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if _, err := io.Copy(tw, fr); err != nil {
			return err
		}

		return nil
	})
}

func Untar(src, dst string) error {
	fr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		dstPath := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if !IsExist(dstPath) && IsDir(dstPath) {
				if err := CreateDir(dstPath); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if dir := filepath.Dir(dstPath); !IsExist(dir) {
				if err := CreateDir(dir); err != nil {
					return err
				}
			}

			file, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			if _, err = io.Copy(file, tr); err != nil {
				return err
			}

			fmt.Println(dstPath)
			file.Close()
		}
	}
}

func CopyEmbed(assets http.FileSystem, embeddedDir, dst string) error {
	return vfsutil.WalkFiles(assets, embeddedDir, func(path string, info os.FileInfo, rs io.ReadSeeker, err error) error {
		if err != nil {
			return err
		}

		if len(path)+1 < len(embeddedDir) {
			return errors.New("embedded dir is empty")
		}
		targetPath := filepath.Join(dst, path[len(embeddedDir)-1:])

		if info.IsDir() {
			return Mkdir(targetPath)
		}

		data, err := vfsutil.ReadFile(assets, path)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(targetPath, data, os.ModePerm)
	})
}

func GetRealPath(p string) (string, error) {
	ok, err := IsSymLink(p)
	if err != nil {
		return p, err
	}
	if !ok {
		return p, nil
	}

	target, err := os.Readlink(p)
	if err != nil {
		return p, err
	}
	return target, nil
}
