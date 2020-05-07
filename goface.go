package main

import (
	"fmt"
	"image"
	"io/ioutil"

	gohash "github.com/corona10/goimagehash"
	pigo "github.com/esimov/pigo/core"
)

var classifier *pigo.Pigo

func main() {}

// callback 回调处理逻辑.
type callback func(src *image.NRGBA, dets []pigo.Detection) []image.Image

// getArr get the result of DetectFace.
func getArr(src *image.NRGBA, dets []pigo.Detection) []image.Image {
	var r []image.Image
	for _, v := range dets {
		if v.Q > 3.14159 {
			x, y, w := v.Col, v.Row, v.Scale/2
			img := src.SubImage(image.Rect(x-w, y-w, x+w, y+w))
			r = append(r, img)
		}
	}
	return r
}

// initClassifier init the classifier.
func initClassifier() {
	if classifier != nil {
		return
	}
	const finder = "./facefinder"
	cascadeFile, err := ioutil.ReadFile(finder)
	if err != nil {
		fmt.Printf("Error reading the cascade file: %v", err)
	}

	pigo := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = pigo.Unpack(cascadeFile)
	if err != nil {
		fmt.Printf("Error reading the cascade file: %s", err)
	}
}

// DetectFace in a picture.
func DetectFace(img image.Image, cb callback) []image.Image {
	initClassifier()
	if classifier == nil || img == nil {
		fmt.Printf("The classifier or image is nil")
		return nil
	}
	src := pigo.ImgToNRGBA(img)
	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     32,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	angle := 0.0 // cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)
	if cb != nil {
		return cb(src, dets)
	}
	return nil
}

// imageCompare 图片比对算法.
func imageCompare(src *gohash.ImageHash, cmp image.Image) float64 {
	if src != nil {
		hash, _ := gohash.AverageHash(cmp)
		if n, err := src.Distance(hash); err == nil {
			return 1 - float64(n)/64.0
		}
	}
	return 0
}

// AlarmProcess 告警处理单元.
// go build -buildmode=plugin goface.go
func AlarmProcess(dis map[string]interface{}, features []interface{}, arr []image.Image, ids []string, level int) bool {
	initClassifier()
	var levelThresholdMap = map[int]float64{0: 0.8, 1: 0.6, 2: 0.8, 3: 0.9}
	threshold := levelThresholdMap[level]
	if hash, ok := dis["hash"]; ok { // 图片比对
		for i, a := range arr {
			v, _ := hash.(*gohash.ImageHash)
			for _, img := range DetectFace(a, getArr) {
				f := imageCompare(v, img)
				if f > threshold {
					fmt.Printf("[%s]相似度阈值%f, 触发告警.", ids[i], f)
					return true
				}
				fmt.Printf("[%s]相似度阈值%f, 未触发告警.", ids[i], f)
			}
		}
	} else {
		if img := dis["image"]; img != nil {
			if v, ok := img.(image.Image); ok {
				if hash, err := gohash.AverageHash(v); err == nil {
					dis["hash"] = hash
				}
			}
		}
		fmt.Printf("计算布控图像的特征值=%v.", dis["hash"])
	}
	return false
}
