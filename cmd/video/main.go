package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/bcoryat/demo/pkg/clarifai"
	"github.com/bcoryat/demo/pkg/config"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

var (
	clarifaiService clarifai.Service
	frames          chan []frameInput
	framesData      chan []clarifai.FrameInfo
	rawFrames       chan gocv.Mat
	captureDevice   *gocv.VideoCapture
	scaledH         int
	batchSize       int
	feed            string
	bStrs           []frameInput
	frameInfos      []clarifai.FrameInfo
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UP")
}

func render(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	for {
		fmt.Printf("(render)queued items: %d\n", len(framesData))
		if err := conn.WriteJSON(<-framesData); err != nil {
			return
		}
	}
}

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/ws", render)
	log.Fatal(http.ListenAndServe("localhost:3000", nil))
}

func scaleWidth(w, h, scaledH float64) int {
	return int(math.Floor((w / h) * scaledH))
}

// FrameInput contains the image as a base64 encoded string and other metadata to help maintain order
type frameInput struct {
	time         int64
	scaledImgStr string
	origImgStr   string
}

func pullFrames() {

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := captureDevice.Read(&img); !ok {
			fmt.Println("cannot read next frame")
			fmt.Printf("isOpened %v\n", captureDevice.IsOpened())

		}
		if img.Empty() {
			continue
		}
		rawFrames <- img
	}
}

func getFrames() {
	//var err error
	//var img gocv.Mat
	i := 0
	for {
		img := <-rawFrames

		scaledW := scaleWidth(float64(img.Size()[1]), float64(img.Size()[0]), float64(scaledH))
		originalImgBuf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
		if err != nil {
			fmt.Println("Error encoding img")
		}
		gocv.Resize(img, &img, image.Point{scaledW, scaledH}, 0, 0, gocv.InterpolationArea)
		buf, err := gocv.IMEncode(gocv.PNGFileExt, img)

		if err != nil {
			fmt.Println("Error encoding img")
		}
		bStrs = append(bStrs, frameInput{time.Now().UnixNano(), base64.StdEncoding.EncodeToString(buf), base64.StdEncoding.EncodeToString(originalImgBuf)})

		i++

		if i == batchSize {
			frames <- bStrs
			bStrs = make([]frameInput, 0)
			i = 0
		}
	}
}

func process(fi frameInput, wg *sync.WaitGroup) {
	start := time.Now()
	fmt.Printf("started Goroutine %d at time %s\n", fi.time, start.Format(time.RFC3339))
	frInfo, err := clarifaiService.PredictByBytes(fi.time, fi.scaledImgStr, fi.origImgStr)
	end := time.Now()
	if err != nil {
		fmt.Println("clarifai service error:", err)
	}
	fmt.Printf("Goroutine %d at time %v  ended\n", fi.time, end.Format(time.RFC3339))
	fmt.Println("go routine elaspesd time: ", time.Since(start))
	if frInfo == nil {
		wg.Done()
	} else {
		frameInfos = append(frameInfos, *frInfo)
		wg.Done()
	}
}

func doBatches() {
	for {
		frInputs := <-frames
		var wg sync.WaitGroup
		for _, fi := range frInputs {
			wg.Add(1)
			go process(fi, &wg)
		}
		wg.Wait()
		sort.Slice(frameInfos, func(i, j int) bool { return frameInfos[i].InputID < frameInfos[j].InputID })
		framesData <- frameInfos
		frameInfos = make([]clarifai.FrameInfo, 0)
		fmt.Println("All go routines finished executing")
		fmt.Printf("(doBatches)queued items: %d\n", len(framesData))
	}
}

func main() {
	frameInfos = make([]clarifai.FrameInfo, 0)

	cfg, err := config.New()
	if err != nil {
		log.Panicln("Config error", err)
	}

	batchSize = cfg.BatchSize
	scaledH = cfg.ScaleHeight

	clarifaiService = clarifai.NewService(cfg.Clarifai.APIKey, cfg.Clarifai.ModelURL)

	rawFrames = make(chan gocv.Mat, 1000)
	frames = make(chan []frameInput, 0)
	framesData = make(chan []clarifai.FrameInfo, 5)

	feed := cfg.RtspFeed

	captureDevice, err = gocv.OpenVideoCapture(feed)

	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", feed)
		return
	}
	defer captureDevice.Close()
	time.Sleep(5)

	go pullFrames()
	go getFrames()
	go doBatches()
	setupRoutes()
}
