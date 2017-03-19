// Mind is a server that streams numbers.
//
// The server is a generator of numbers, which are sent to a client
// via the websocket protocol. The client, which is some HTML and JS
// returned by visiting the index page, visualizes the stream of numbers.
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Mind struct {
	x int
}

const IndexHTML = `
<html>
<head>
<script>
var exampleSocket = new WebSocket("ws://localhost:8080/stream");

exampleSocket.onmessage = function (event) {
  d = event.data
  var c = document.getElementById("myCanvas");
  var ctx = c.getContext("2d");

  ctx.fillStyle = 'rgb('+d+', '+d+', '+d+')';
  ctx.fillRect(10, 10, 500, 500);
}
exampleSocket.onopen = function (event) {
  //  console.log("Sending some stuff to server");
  exampleSocket.send("Here's some text that the server is urgently awaiting!"); 
};
</script>
</head>
<body>
<div>
<canvas id="myCanvas" width="1000" height="1000" style="border:1px solid #000000;">
</canvas>
<p>Woop</p>
<p>Woop</p>

</div>
</body>
`

func f(x int) int {
	max := 255
	return int(float64(x)*rand.Float64()+rand.Float64()*50) % max
}

func (m *Mind) change() {
	m.x = f(m.x)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request")
	fmt.Fprintf(w, IndexHTML)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got ws request")
	u := websocket.Upgrader{}
	conn, err := u.Upgrade(w, r, http.Header{})
	if err != nil {
		log.Printf("failed to upgrade http request to websocket: %v\n", err)
		return
	}
	log.Println("upgraded to websocket")
	m := Mind{1000}
	for {
		time.Sleep(time.Millisecond * 100)
		log.Printf("%+v: %q\n", m, string(m.x*'.'))
		m.change()
		msg := []byte(fmt.Sprintf("%d", m.x))
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("failed to fetch writer from websocket: %v\n", err)
			return
		}
	}
}

func main() {
	log.Println("Connect to websocket to hear your mind")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/stream", wsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
