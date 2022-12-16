package main

// To Do:
// - Implement Error handling
// https://earthly.dev/blog/golang-errors/
// - Naming: naming follow frame number
// - write program for detecting camera-cuts
// - concurrency in exracting frames?
// - Sobel-implementation
// - Modul: Camera switcher

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/wielandos/sobelfilter"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gocv.io/x/gocv"
)

var imgSobel image.Image

type MyImg struct {
	// Embed image.Image so MyImg will implement image.Image
	// because fields and methods of Image will be promoted:
	image.Image
}

func main() {

	/*
		ExtractFrames("./video_raw/1.mp4", "./frames_raw/", 324, 327)
	*/

	imgPath := "./frames_raw/out8133.jpeg"
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	//img := gocv.IMRead(imgPath, gocv.IMReadGrayScale)
	if img.Empty() {
		fmt.Printf("Failed to read image: %s\n", imgPath)
		os.Exit(1)
	}
	// Convert BGR to HSV image (dont modify the original)
	hsvImg := ConvertToHSV(img)
	SaveFile("./frames_proc/hsv_1961.jpg", hsvImg)

	// Convert to grey:
	greyImg := gocv.NewMat()
	gocv.CvtColor(img, &greyImg, gocv.ColorBGRToGray)
	SaveFile("./frames_proc/grey_1961.jpg", greyImg)

	// Blur:
	BlurProcessed := gocv.NewMat()
	gocv.GaussianBlur(img, &BlurProcessed, image.Pt(3, 3), 0, 0, gocv.BorderDefault)
	SaveFile("./frames_proc/blur_1961.jpg", BlurProcessed)

	imgSobel = InputSobel("./frames_proc/blur_1961.jpg")
	imgSobel = sobelfilter.ApplySobelFilter(imgSobel) //converts "img" to grayscale and runs edge detect. Returns an image.Image with changes.
	SaveSobel("./frames_proc/sobel_6.jpg", imgSobel, 180)

	// rotate
	RotateImg := gocv.NewMat()
	gocv.Rotate(BlurProcessed, &RotateImg, gocv.Rotate90CounterClockwise)
	SaveFile("./frames_proc/rotate_1961.jpg", RotateImg)

	f, err := os.Open("./frames_proc/rotate_1961.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img_, err := drawableRGBImage(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Model:", img_.ColorModel())
	fmt.Println("Bounds:", img_.Bounds())
	fmt.Println("At(1,2):", img_.At(1, 2))
	img_.Set(1, 2, color.White)
	fmt.Println("At(1,2):", img_.At(1, 2), "(after Set)")

}

func (m *MyImg) At(x, y int) color.Color {
	// "Changed" part: custom colors for specific coordinates:
	switch {
	case x == 0 && y == 0:
		return color.RGBA{85, 165, 34, 255}
	case x == 0 && y == 1:
		return color.RGBA{255, 0, 0, 255}
	}
	// "Unchanged" part: the colors of the original image:
	return m.Image.At(x, y)
}

func drawableRGBImage(f io.Reader) (draw.Image, error) {
	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	output_rgb := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(output_rgb, output_rgb.Bounds(), img, b.Min, draw.Src)

	return output_rgb, nil
}

// Get the bi-dimensional pixel array
func getPixels(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

// Pixel struct example
type Pixel struct {
	R int
	G int
	B int
	A int
}

func InputSobel(inputPath string) image.Image {
	f, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	imgInputSobel, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return imgInputSobel
}

func SaveSobel(targetPath string, img image.Image, quality int) {
	e, err := os.Create(targetPath)
	if err != nil {
		fmt.Println("exist already")
	}
	defer e.Close()
	opt := jpeg.Options{
		Quality: quality,
	}
	err = jpeg.Encode(e, img, &opt)
	if err != nil {
		// Handle error
	}

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
