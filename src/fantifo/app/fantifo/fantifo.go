package fantifo

import (
	"container/list"
	"time"
	"os"
	"image/color"
	"image/jpeg"
	"image/png"
	"fmt"
	"io/ioutil"
	"image"
	"strings"
	"errors"
)

type Tick struct {
	Tick int
	Data *BufferedImage
	Time int64
}

type Subscription struct {
	New <-chan Tick // New events coming in.
}

type BufferedImage struct {
	Image []color.Color
}

const (
	tickFrequency time.Duration = 100
	tickDelay int64 = 1000
)

var (
	images []BufferedImage
	subscribe = make(chan (chan <- Subscription), 10)
// Send a channel here to unsubscribe.
	unsubscribe = make(chan (<-chan Tick), 10)
// Send events here to publish them.
	publish = make(chan Tick, 10)
)

// Owner of a subscription must cancel it when they stop listening to events.
func (s Subscription) Close() {
	unsubscribe <- s.New // Unsubscribe the channel.
	drain(s.New)         // Drain it, just in case there was a pending publish.
}

func Subscribe() Subscription {
	resp := make(chan Subscription)
	subscribe <- resp
	return <-resp
}

// This function loops forever, handling the chat room pubsub
func fantifo() {
	subscribers := list.New()

	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Tick, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}

		case event := <-publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Tick) <- event
			}

		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Tick) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		}
	}
}

func readImage(path string) ([]color.Color, error) {
	file, _ := os.Open(path)
	defer file.Close()

	var img image.Image
	var imgErr error = errors.New("No decoder available")

	if strings.HasSuffix(file.Name(), ("jpg")) {
		img, imgErr = jpeg.Decode(file)
	} else if strings.HasSuffix(file.Name(), ("png")) {
		img, imgErr = png.Decode(file)
	}

	if imgErr != nil {
		fmt.Println(imgErr)
		return nil, imgErr
	}

	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()

	decodedImage := make([]color.Color, dx * dy, dx * dy)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			decodedImage[x + y * dy] = img.At(x, y)
		}
	}

	return decodedImage, nil
}

func readImages() {
	directory := "src/fantifo/assets/"
	files, _ := ioutil.ReadDir(directory)

	images = make([]BufferedImage, len(files))
	i := 0

	for _, f := range files {
		path := directory + f.Name()
		img, err := readImage(path)
		if err != nil {
			fmt.Println("Unable to load image: " + path)
			fmt.Println(err)
		} else {
			images[i] = BufferedImage{img}
			i++
		}
	}
}

func makeFakeImage(index int) []color.Color {
	dx := 12
	dy := 12
	decodedImage := make([]color.Color, dx * dy, dx * dy)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			cIndex := x + y * dx
			if x == index % dx {
				decodedImage[cIndex] = color.RGBA{255, 0, 0,1}
			} else {
				decodedImage[cIndex] = color.RGBA{0, 0, 0,1}
			}
		}
	}

	return decodedImage
}

func fakeReadImages() {
	count := 144
	images = make([]BufferedImage, count)

	for index := 0; index < count; index++ {
		img := makeFakeImage(index)
		images[index] = BufferedImage{img}
	}
}

func tick() {
	ticker := time.NewTicker(time.Millisecond * tickFrequency)
	counter := 0

	for _ = range ticker.C {
		image := images[counter % len(images)]

		if len(image.Image) > 0 {
			publish <- MakeTick(counter, &image)
		}

		counter++;
	}
}

//////////////
// Helpers ///
//////////////

// Drains a given channel of any messages.
func drain(ch <-chan Tick) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

// Make new tick with default parameters
func MakeTick(tick int, data *BufferedImage) Tick {
	return Tick{tick, data, MakeTimestamp() + tickDelay}
}

// Get current time in milliseconds
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

///////////////
//   Main   ///
///////////////

func init() {
	fakeReadImages()
	//readImages()
	go fantifo()
	go tick()
}