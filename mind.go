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
	"unicode"

	"github.com/gorilla/websocket"
)

type Mind struct {
	x int
}

const IndexHTML = `
<!DOCTYPE html>
<head>
<style>
div#wrapper {
  height: 800px;
  width: 800px;
  margin: auto;
}

canvas {}
</style>
</head>
<body>
<div id="wrapper">
<canvas id="myCanvas" width="800" height="800" style="border:1px solid #d3d3d3;">
Your browser does not support the HTML5 canvas tag.
</canvas>
</div>

<script>
var exampleSocket = new WebSocket("ws://localhost:8080/stream");
var t=10;

exampleSocket.onmessage = function (event) {
  var d = event.data;
  var c = document.getElementById("myCanvas");
  var ctx = c.getContext("2d");

  ctx.fillStyle = 'rgb('+d+', '+d+', '+d+')';

  ctx.fillRect(t, 10, 1, 780);
  t = t + 1;
  if (t > 790) {
    t = 10;
  }

}
exampleSocket.onopen = function (event) {
  //  console.log("Sending some stuff to server");
  exampleSocket.send("Here's some text that the server is urgently awaiting!"); 
};
</script>

</body>
`

func g(x int) int {
	//	return int(rand.Float64() * max)
	max := 255
	return int(float64(x)*rand.Float64()+rand.Float64()*50) % max
}

func f(x int) int {
	d := rand.Float64() - rand.Float64()
	return int(float64(x)+d) % 255
}

func (m *Mind) change() {
	m.x = g(m.x)
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
	i := 0
	runes := ""
	for {
		if i%100 == 0 {
			if i%10000 == 0 {
				log.Println(runes)
			}
			runes = ""
		}
		time.Sleep(time.Nanosecond * 1000 * 10)
		m.change()
		r := m.x * '.'
		if unicode.IsPrint(rune(r)) {
			runes += string(r)
		}
		msg := []byte(fmt.Sprintf("%d", m.x))
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("failed to fetch writer from websocket: %v\n", err)
			return
		}
		i += 1
	}
}

func main() {
	log.Println("Connect to websocket to hear your mind")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/stream", wsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
