package help

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// 判断路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 清空目录
func CleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer func() {
		err := d.Close()
		if err != nil {
			log.Fatalf("error clean dir: %s", err)
		}
	}()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// dir 目录
// extension 文件后缀
func WalkDir(dir string, extension map[string]int) (map[string]string, error) {
	var fileNames = make(map[string]string)

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				suffix := filepath.Ext(path)
				if _, exists := extension[suffix]; exists {
					rel, err := filepath.Rel(dir, path)
					if err != nil {
						return filepath.SkipDir
					}
					fileNames[rel] = path
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return fileNames, nil
}

func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s func %v\n", what, time.Since(start))
	}
}
