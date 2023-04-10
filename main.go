package main

// To Do:
// - Implement Error handling
// https://earthly.dev/blog/golang-errors/
// - Naming: naming follow frame number
// - insert hierachical clustering for detecting camera-cuts
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
	"math"
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

/*
	type Drawing interface {
		rectangle() string
	}

// CustomImage is embedded struct: means that we can add a nested struct
// and access it more easily
// => in the present case: we use a struct from the another package

func (img CustomImage) rectangle() string {

	return "test"

}
*/
type CustomImage struct {
	draw.Image
}

type PositionsRectange interface {
	Position() [2]int
	Size() int
}

type Rect struct {
	height, width int
}

type Img struct {
	size []image.Point
}

// Pixel struct example
type Pixel struct {
	R int
	G int
	B int
	A int
}

type positions interface {
	position() []int
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
	//gocv.Rotate(BlurProcessed, &RotateImg, gocv.Rotate90CounterClockwise)
	gocv.Rotate(BlurProcessed, &RotateImg, gocv.Rotate180Clockwise)
	SaveFile("./frames_proc/rotate_1961.jpg", RotateImg)

	f, err := os.Open("./frames_proc/blur_1961.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img_, _ := drawableRGBImage(f)
	custImg := CustomImage{
		img_,
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("type: %T\n", img_)
	fmt.Println("Model:", img_.ColorModel())
	fmt.Println("Bounds:", img_.Bounds())
	fmt.Println("At(1,2):", img_.At(1, 2))
	fmt.Println("At(1,2):", img_.At(13, 10), "(after Set)")

	out, err := os.Create("./frames_proc/rectagnle_1961.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	myRectangle := image.Rect(0, 260, 1100, 120)

	dst := addRectangle(custImg, myRectangle)

	err = jpeg.Encode(out, dst, nil) // put quality to 80%
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// read image
	img_rec := gocv.IMRead("./frames_proc/rectagnle_1961.jpg", gocv.IMReadColor)
	if img_rec.Empty() {
		fmt.Printf("Failed to read image: %s\n", img_rec)
		os.Exit(1)
	}

	// origImg := []image.Point{
	// 		image.Point{128, 165}, // top-left
	// 		image.Point{215, 275}, // bottom-left
	// 		image.Point{385, 128}, // bottom-right
	// 		image.Point{300, 40},  // top-right
	// 	}

	// image coordinages corners of the select business card object
	var origImg Img

	origImg.size = []image.Point{
		image.Point{10, 190},   // top-left
		image.Point{10, 240},   // bottom-left
		image.Point{1000, 200}, // bottom-right
		image.Point{1000, 150}, // top-right
	}

	// Add Point Slice
	out_marked, err := os.Create("./frames_proc/point_1961.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	img_marked := addPointVector(custImg, origImg.size)

	err = jpeg.Encode(out_marked, img_marked, nil) // put quality to 80%
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// calculate height as a distance between (top-left, bottom-left) and (top-right, bottom-right)
	heightA := math.Sqrt(math.Pow(float64(origImg.size[0].X-origImg.size[1].X), 2) + math.Pow(float64(origImg.size[0].Y-origImg.size[1].Y), 2))
	heightB := math.Sqrt(math.Pow(float64(origImg.size[3].X-origImg.size[2].X), 2) + math.Pow(float64(origImg.size[3].Y-origImg.size[2].Y), 2))
	height := int(math.Max(heightA, heightB))

	// caluclate width as a distance between () and ()
	widthA := math.Sqrt(math.Pow(float64(origImg.size[0].X-origImg.size[3].X), 2) + math.Pow(float64(origImg.size[0].Y-origImg.size[3].Y), 2))
	widthB := math.Sqrt(math.Pow(float64(origImg.size[1].X-origImg.size[2].X), 2) + math.Pow(float64(origImg.size[1].Y-origImg.size[2].Y), 2))
	width := int(math.Max(widthA, widthB))
	/*
		newImg := []image.Point{
			image.Point{0, 0},
			image.Point{0, height},
			image.Point{width, height},
			image.Point{width, 0},
		}
	*/
	var newImg Img

	newImg.size = []image.Point{
		image.Point{0, 0},
		image.Point{0, height},
		image.Point{width, height},
		image.Point{width, 0},
	}

	fmt.Println(newImg)
	src := gocv.NewPointVectorFromPoints(origImg.size)
	dest := gocv.NewPointVectorFromPoints(newImg.size)

	fmt.Println(src)
	transform := gocv.GetPerspectiveTransform(src, dest)
	perspective := gocv.NewMat()
	gocv.WarpPerspective(img_rec, &perspective, transform, image.Point{width, height})
	//gocv.WarpPerspective(img_rec, &img_rec, transform, image.Point{width, height})

	//outPath := "card_perspective.jpg"
	if ok := gocv.IMWrite("card_perspective.jpg", perspective); !ok {
		fmt.Printf("Failed to write image: %s\n")
		os.Exit(1)
	}

}

func (img Img) position() [2]int {

	widthA := math.Sqrt(math.Pow(float64(img.size[0].X-img.size[3].X), 2) + math.Pow(float64(img.size[0].Y-img.size[3].Y), 2))
	widthB := math.Sqrt(math.Pow(float64(img.size[1].X-img.size[2].X), 2) + math.Pow(float64(img.size[1].Y-img.size[2].Y), 2))
	width := int(math.Max(widthA, widthB))

	heightA := math.Sqrt(math.Pow(float64(img.size[0].X-img.size[1].X), 2) + math.Pow(float64(img.size[0].Y-img.size[1].Y), 2))
	heightB := math.Sqrt(math.Pow(float64(img.size[3].X-img.size[2].X), 2) + math.Pow(float64(img.size[3].Y-img.size[2].Y), 2))
	height := int(math.Max(heightA, heightB))
	pos := [2]int{width, height}
	return pos

}

// how do I implement this function if every method of the interface has different return type?
func extract(p positions) []int {
	fmt.Println(p)
	fmt.Println(p.position())
	return p.position()
}

func (r Rect) Size() int {
	return r.height * r.width

}

func addRectangle(img CustomImage, rect image.Rectangle) draw.Image {
	myColor := color.RGBA{255, 0, 255, 255}

	min := rect.Min
	max := rect.Max

	for i := min.X; i < max.X; i++ {
		img.Set(i, min.Y, myColor)
		img.Set(i, max.Y, myColor)
	}

	for i := min.Y; i <= max.Y; i++ {
		img.Set(min.X, i, myColor)
		img.Set(max.X, i, myColor)
	}
	return img
}

func addPointVector(img CustomImage, pointSlice []image.Point) draw.Image {

	for i, _ := range pointSlice {
		fmt.Println(i)
		point := pointSlice[i]
		myRectangle := image.Rect(point.X, point.Y, point.X-10, point.Y-10)

		addRectangle(img, myRectangle)
	}
	return img
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
