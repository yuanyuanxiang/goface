package goface

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"

	"github.com/corona10/goimagehash"
	pigo "github.com/esimov/pigo/core"
)

var classifier *pigo.Pigo

var saveFlag bool

// save the result of DetectFace.
func save(src *image.NRGBA, dets []pigo.Detection) {
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
}

// init the classifier.
func init() {
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
func DetectFace(img image.Image) {
	if classifier == nil || img == nil {
		fmt.Printf("The classifier or image is nil")
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
	if saveFlag {
		save(src, dets)
	}
}

// imageCompare 图片比对算法.
func imageCompare(src *goimagehash.ImageHash, cmp image.Image) float64 {
	if src != nil {
		hash, _ := goimagehash.AverageHash(cmp)
		if n, err := src.Distance(hash); err == nil {
			return 1 - float64(n)/64.0
		}
	}
	return 0
}

// AlarmProcess 告警处理单元.
func AlarmProcess(dis map[string]interface{}, features []interface{}, arr []image.Image, ids []string, level int) bool {
	var levelThresholdMap = map[int]float64{0: 0.8, 1: 0.6, 2: 0.8, 3: 0.9}
	threshold := levelThresholdMap[level]
	if hash, ok := dis["hash"]; ok { // 图片比对
		for i, a := range arr {
			v, _ := hash.(*goimagehash.ImageHash)
			f := imageCompare(v, a)
			if f > threshold {
				fmt.Printf("[%s]相似度阈值%f, 触发告警.", ids[i], f)
				return true
			}
			fmt.Printf("[%s]相似度阈值%f, 未触发告警.", ids[i], f)
		}
	} else {
		if img := dis["image"]; img != nil {
			if v, ok := img.(image.Image); ok {
				if hash, err := goimagehash.AverageHash(v); err == nil {
					dis["hash"] = hash
				}
			}
		}
		fmt.Printf("计算布控图像的特征值=%v.", dis["hash"])
	}
	return false
}
