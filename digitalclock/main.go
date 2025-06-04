//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const TimeParamKey = "time"
const TimeLayout = "15:04:05"
const CorrectTimeLen = 8

const ScaleParamKey = "k"
const DefaultScaleParam = "1"
const MinScaleValue = 1
const MaxScaleValue = 30

const ContentTypeHeader = "Content-Type"
const ImageContentType = "image/png"

var InvalidTimeMsg = []byte("invalid time")
var InvalidScaleMsg = []byte("invalid k")

func TimeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	timeParam := params.Get(TimeParamKey)
	if !params.Has(TimeParamKey) {
		timeParam = time.Now().Format(TimeLayout)
	}
	if len(timeParam) != CorrectTimeLen {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(InvalidTimeMsg)
		return
	}
	timeValue, err := time.Parse(TimeLayout, timeParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(InvalidTimeMsg)
		return
	}

	scaleParam := params.Get(ScaleParamKey)
	if !params.Has(ScaleParamKey) {
		scaleParam = DefaultScaleParam
	}
	scaleValue, err := strconv.Atoi(scaleParam)
	if err != nil || scaleValue < MinScaleValue || scaleValue > MaxScaleValue {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(InvalidScaleMsg)
		return
	}

	img := generateImg(timeValue, scaleValue)

	w.Header().Set(ContentTypeHeader, ImageContentType)
	if err := png.Encode(w, img); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "port to run server on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", TimeHandler)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func generateImg(timeValue time.Time, k int) image.Image {
	height := calcExpectedHeight(k)
	width := calcExpectedWidth(k)
	widthDigit := calcSymbolWidth(Zero) * k
	widthColon := calcSymbolWidth(Colon) * k

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := range width {
		for y := range height {
			img.Set(x, y, color.White)
		}
	}

	h := timeValue.Hour()
	h0 := h / 10
	h1 := h % 10
	printSymbol(scaleSymbol(getSymbol(h0), k), height, widthDigit, 0, img)
	printSymbol(scaleSymbol(getSymbol(h1), k), height, widthDigit, widthDigit, img)
	printSymbol(scaleSymbol(Colon, k), height, widthColon, 2*widthDigit, img)

	m := timeValue.Minute()
	m0 := m / 10
	m1 := m % 10
	printSymbol(scaleSymbol(getSymbol(m0), k), height, widthDigit, 2*widthDigit+widthColon, img)
	printSymbol(scaleSymbol(getSymbol(m1), k), height, widthDigit, 3*widthDigit+widthColon, img)
	printSymbol(scaleSymbol(Colon, k), height, widthColon, 4*widthDigit+widthColon, img)

	s := timeValue.Second()
	s0 := s / 10
	s1 := s % 10
	printSymbol(scaleSymbol(getSymbol(s0), k), height, widthDigit, 4*widthDigit+2*widthColon, img)
	printSymbol(scaleSymbol(getSymbol(s1), k), height, widthDigit, 5*widthDigit+2*widthColon, img)

	return img
}

func printSymbol(sym string, height int, width int, shift int, img *image.RGBA) {
	sym = strings.ReplaceAll(sym, "\n", "")
	for y := range height {
		for x := range width {
			if sym[y*width+x] == '1' {
				img.Set(shift+x, y, Cyan)
			}
		}
	}
}

func scaleSymbol(s string, k int) string {
	width := len(strings.SplitN(s, "\n", 2)[0])
	height := len(strings.Split(s, "\n"))
	scaledWidth := k * width
	scaledHeight := k * height

	s = strings.ReplaceAll(s, "\n", "")
	scaled := make([]byte, scaledWidth*scaledHeight)

	for w := range width {
		for h := range height {
			for i := range k {
				for j := range k {
					wN := k*w + i
					hN := k*h + j
					scaled[scaledWidth*hN+wN] = s[width*h+w]
				}
			}

		}
	}
	return string(scaled)
}

func getSymbol(i int) string {
	switch i {
	case 0:
		return Zero
	case 1:
		return One
	case 2:
		return Two
	case 3:
		return Three
	case 4:
		return Four
	case 5:
		return Five
	case 6:
		return Six
	case 7:
		return Seven
	case 8:
		return Eight
	case 9:
		return Nine
	default:
		panic(fmt.Sprintf("%v is not supported digit", i))
	}
}

func calcSymbolWidth(s string) int {
	return len(strings.SplitN(s, "\n", 2)[0])
}

func calcExpectedWidth(k int) int {
	return (calcSymbolWidth(Zero)*6 + calcSymbolWidth(Colon)*2) * k
}

func calcExpectedHeight(k int) int {
	return len(strings.Split(Zero, "\n")) * k
}
