package controllers

import (
	"golang.org/x/net/websocket"
	"github.com/revel/revel"
	"fantifo/app/fantifo"
	"strconv"
	"encoding/json"
	"strings"
	"image/color"
)

const (
	LATENCY_TEST int = 1
	REGISTER_CLIENT int = 2
	TICK int = 3
	DATA int = 4
)

type latencyTestRequest struct {
	Type      int
	TimeStamp int64
}

type latencyTestResponse struct {
	Type       int
	ServerTimeStamp int64
	TimeStamp  int64
}

type clientRegisterResponse struct {
	Type int
	Id   int
}

type dataResponse struct {
	Type  int
	Color color.RGBA
	Time  int64
}

type Fantifo struct {
	*revel.Controller
}

func (c Fantifo) Endpoint(user string) revel.Result {
	return c.Render(user)
}

func (c Fantifo) EndpointSocket(user string, ws *websocket.Conn) revel.Result {
	position, _ := strconv.Atoi(user)
	subscription := fantifo.Subscribe()
	defer subscription.Close()

	// Listen for client messages
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				return
			}

			// Extract a number between 1 - 9 from the first position of the request, ex "1{JSON...}"
			messageType, _ := strconv.Atoi(string(msg[0]))
			message := strings.TrimPrefix(msg, string(msg[0]))

			if (messageType == LATENCY_TEST) {
				eveIn := latencyTestRequest{}
				json.Unmarshal([]byte(message), &eveIn)

				timeOffset := fantifo.MakeTimestamp()
				eventType := eveIn.Type
				timeStamp := eveIn.TimeStamp
				eveOut := latencyTestResponse{eventType, timeOffset, timeStamp}

				websocket.JSON.Send(ws, eveOut)
			}
		}
	}()

	// Listen for new events from the main loop
	for {
		select {

		// Events originate in main event loop
		case event := <-subscription.New:
			index := position % len(event.Data.Image)

		// Convert 16 bit int to 8 bit byte
			r0, g0, b0, a0 := event.Data.Image[index].RGBA()
			r := uint8(r0 / 256);
			g := uint8(g0 / 256);
			b := uint8(b0 / 256);
			a := uint8(a0 / 256);

			data := color.RGBA{r, g, b, a}
			timeStamp := event.Time

			response := dataResponse{DATA, data, timeStamp}
			if websocket.JSON.Send(ws, &response) != nil {
				// Client disconnected.
				return nil
			}
		}
	}
	return nil
}
