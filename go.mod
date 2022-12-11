module github.com/wielandos/AdGang

go 1.19

require (
	github.com/disintegration/imaging v1.6.2
	github.com/u2takey/ffmpeg-go v0.4.1
	github.com/wielandos/sobelfilter v0.0.0-00010101000000-000000000000
	gocv.io/x/gocv v0.31.0
)

require (
	github.com/aws/aws-sdk-go v1.38.20 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/u2takey/go-utils v0.3.1 // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
)

replace github.com/wielandos/sobelfilter => ../sobelfilter
