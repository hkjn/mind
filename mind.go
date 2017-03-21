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
  height: 1024px;
  width: 1024px;
  margin: auto;
}

canvas {}
</style>
</head>
<body>
<div id="wrapper">
<canvas id="myCanvas" width="1024" height="1024" style="border:1px solid #d3d3d3;">
Your browser does not support the HTML5 canvas tag.
</canvas>
</div>

<script>
var exampleSocket = new WebSocket("ws://localhost:8080/stream");
var t=0;

exampleSocket.onmessage = function (event) {
  var d = parseInt(event.data);
  var c = document.getElementById("myCanvas");
  var ctx = c.getContext("2d");
  var r = g = b = d;
  r += 1 * t;
  r = parseInt(r % 255);
  g += 1.1 * t;
  g = parseInt(g % 255);
  b += 1.2 * t;
  b = parseInt(b % 255);
  if (t == 0) {
    console.log(d, r, g, b);
  }
  ctx.fillStyle = 'rgb('+
      r + ', ' +
      g + ', ' +
      b + ')';

  ctx.fillRect(t, 0, 1, 255);
  t = t + 1;
  if (t > 255) {
    t = 0;
  }

}
exampleSocket.onopen = function (event) {
  //  console.log("Sending some stuff to server");
  var c = document.getElementById("myCanvas");
  var ctx = c.getContext("2d");
  ctx.scale(4, 4);
  exampleSocket.send("Here's some text that the server is urgently awaiting!"); 
};
</script>

</body>
`

const Max = 255

func r(x int) int {
	return int(rand.Float64() * Max)
}

func h(x int) int {
	d := 1 - (rand.Float64()*0.5 - rand.Float64()*0.5)
	return int(float64(x)*d) % Max
}

func f(x int) int {
	return (x + 3) % Max
}

func (m *Mind) change() {
	m.x = r(m.x)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request")
	fmt.Fprintln(w, IndexHTML)
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
		time.Sleep(time.Nanosecond * 1000 * 100)
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
