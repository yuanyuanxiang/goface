package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"plugin"
	"testing"

	pigo "github.com/esimov/pigo/core"
)

// save the result of DetectFace.
func save(src *image.NRGBA, dets []pigo.Detection) []image.Image {
	for i, v := range dets {
		x, y, w := v.Col, v.Row, v.Scale/2
		img := src.SubImage(image.Rect(x-w, y-w, x+w, y+w))
		file, err := os.Create(fmt.Sprintf("%d-%.2f", i, v.Q) + ".jpg")
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer file.Close()
		if err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100}); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

// https://github.com/esimov/pigo
func TestGoface(t *testing.T) {
	file, err := os.Open("./test.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println(err)
	}
	DetectFace(img, save)
}

// dis-布控任务；features-特征值列表；arr-子图列表；ids-子图id列表；level-预警灵敏度.
type alarmCallback func(dis map[string]interface{}, features []interface{}, arr []image.Image, ids []string, level int) bool

// loadPlugin 加载第三方动态库.该动态库实现了形如"AlarmCallback"的告警处理函数.
func loadPlugin(soPath, funName string) (alarmCallback, error) {
	// 打开so文件
	p, err := plugin.Open(soPath)
	if err != nil {
		return nil, fmt.Errorf("open failed(%v)", err)
	}
	// 查找函数
	f, err := p.Lookup(funName)
	if err != nil {
		return nil, fmt.Errorf("lookup failed(%v)", err)
	}
	// 转换类型后调用函数
	if cb, ok := f.(func(map[string]interface{}, []interface{}, []image.Image, []string, int) bool); ok {
		return cb, nil
	}
	return nil, fmt.Errorf("'%s' has no Alarm func", soPath)
}

// readPluginsDir 读取动态链接库所在目录，以便加载第三方动态链接库.
func readPluginsDir(dir string) []alarmCallback {
	var fun []alarmCallback
	files, err := ioutil.ReadDir(dir)
	fmt.Printf("readPluginsDir in '%v'.\n", dir)
	if err != nil {
		return fun
	}
	for _, f := range files {
		name := f.Name()
		if f.IsDir() || len(name) <= 3 || name[len(name)-3:] != ".so" {
			continue
		}
		cur := dir + "/" + name
		if cb, err := loadPlugin(cur, "AlarmProcess"); err == nil {
			fun = append(fun, cb)
			fmt.Printf("load '%s' succeed.\n", cur)
		} else {
			fmt.Printf("load '%s' failed, err= %v.\n", cur, err)
		}
	}
	return fun
}

// test load "goface.so".
func TestLoadGoface(t *testing.T) {
	path := "./test.jpg"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	var dis = map[string]interface{}{"image": img}
	var arr = []image.Image{img}
	var ids = []string{path}

	alarms := readPluginsDir("./")

	for i := 0; i < 10; i++ {
		for _, fun := range alarms {
			fun(dis, nil, arr, ids, 2)
		}
	}
}
