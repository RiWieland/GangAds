package main

// To Do:
// - Implement Error handling
// https://earthly.dev/blog/golang-errors/

// - write program for detecting camera-cuts

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"strconv"

	"github.com/disintegration/imaging"

	//"reflect"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gocv.io/x/gocv"
)

var edge image.Image

func main() {

	f, err := os.Open("./frames_proc/blur_1.jpg")
	//f, err := os.Open("/Users/richardwieland/Desktop/Projects/AdCoVi/frames_raw/carre_homme.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	imgEdge, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	edge = ApplyFilter(imgEdge) //converts "img" to grayscale and runs edge detect. Returns an image.Image with changes.

	e, err := os.Create("./frames_proc/sobel_4.jpg")
	if err != nil {
		fmt.Println("exist already")
	}
	defer e.Close()
	opt := jpeg.Options{
		Quality: 90,
	}
	err = jpeg.Encode(e, edge, &opt)
	if err != nil {
		// Handle error
	}

	//////
	imgPath := "/Users/richardwieland/Desktop/Projects/AdCoVi/frames_raw/out1960.jpeg"
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	//img := gocv.IMRead(imgPath, gocv.IMReadGrayScale)
	if img.Empty() {
		fmt.Printf("Failed to read image: %s\n", imgPath)
		os.Exit(1)
	}
	// Convert BGR to HSV image (dont modify the original)
	hsvImg := ConvertToHSV(img)
	SaveFile("./frames_proc/hsv_1.jpg", hsvImg)

	// Convert to grey:
	greyImg := gocv.NewMat()
	gocv.CvtColor(img, &greyImg, gocv.ColorBGRToGray)
	SaveFile("./frames_proc/grey_1.jpg", greyImg)

	// Blur:
	BlurProcessed := gocv.NewMat()
	gocv.GaussianBlur(greyImg, &BlurProcessed, image.Pt(3, 3), 0, 0, gocv.BorderDefault)
	SaveFile("./frames_proc/blur_1.jpg", BlurProcessed)

}

func ConvertToHSV(image gocv.Mat) gocv.Mat {
	hsvImg := gocv.NewMat()
	gocv.CvtColor(image, &hsvImg, gocv.ColorRGBToHSV)

	return hsvImg
}

func SaveFile(targetPath string, image gocv.Mat) {
	if ok := gocv.IMWrite(targetPath, image); !ok {
		fmt.Printf("Failed to write image:")
		os.Exit(1)
	}
}

/*
func ApplyGaussBlur(image) image {

	gocv.GaussianBlur(immage, &ball, image.Pt(35, 35), 0, 0, gocv.BorderDefault)
	// write image to filesystem
	outPath := filepath.Join("blur_messi.jpg")

}
*/

// input_path = "/Users/richardwieland/Desktop/Projects/AdCoVi/video_raw/1.mp4"
// output_path = "./frames_raw/"
func ExtractFrames(input_path string, output_path string, start_sec int, end_sec int) {
	target_frames := GetFramesPerSec(start_sec, end_sec)

	for i := target_frames[0]; i < target_frames[1]; i++ {

		reader := ExampleReadFrameAsJpeg(input_path, (int(i)))
		img, err := imaging.Decode(reader)
		if err != nil {
			fmt.Println("ERROR")
		}

		str := strconv.Itoa(i)
		target_path := output_path + "out" + str + ".jpeg"
		err = imaging.Save(img, target_path)
		if err != nil {
			fmt.Println("ERROR")
		}
	}
}

func ExampleReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {

	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

func GetFramesPerSec(startSec int, endSec int) [2]int {
	var FrameArray [2]int
	FrameArray[0] = int(float64(startSec) * 25.1)
	FrameArray[1] = int((float64(endSec) * 25.1))
	// 30, 60 oder gar 120
	return FrameArray
}

/*

1960 : 78 =
fps: 25,1 for vid 1


func GetFramesPerSec(startSec int, endSec int) {
	var FrameArray [2]int
	FrameArray[0] =
	FrameArray[1] =
	// 30, 60 oder gar 120
}
*/
