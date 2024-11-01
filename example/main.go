package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/radenrishwan/aci/chttp"
	"github.com/radenrishwan/aci/cwebsocket"
)

func main() {
	webscoketExample()
}

func httpExample() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	router := chttp.NewRouter()

	router.HandleFunc("/", func(r chttp.Request) *chttp.Response {
		return chttp.NewTextResponse("OK")
	})

	router.HandleFunc("/hello", func(r chttp.Request) *chttp.Response {
		return chttp.NewTextResponse("Hello")
	})

	for {
		conn, _ := server.Accept()

		err := router.Execute(conn)
		if err != nil {
			log.Println(err)
		}

		conn.Close()
	}
}

func webscoketExample() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, _ := server.Accept()

		go func() {
			// upgrade connection to websocket
			err := cwebsocket.Upgrade(conn)
			if err != nil {
				log.Println(err)
				conn.Close()
				return
			}

			for {
				msg, err := cwebsocket.Read(conn)
				if err != nil {
					// conver err into WsError
					err, ok := err.(*cwebsocket.WsError)
					if ok {
						if err.Reason == "EOF" {
							cwebsocket.Close(conn, "", cwebsocket.STATUS_CLOSE_NORMAL_CLOSURE)
						}

						return
					}

					log.Println(err)
					cwebsocket.Close(conn, "", cwebsocket.STATUS_CLOSE_NORMAL_CLOSURE)

					return
				}

				fmt.Print(string(msg))

				// write a websocket frame to the connection
				err = cwebsocket.Write(conn, []byte("Hello, World"))
				if err != nil {
					log.Println(err)
					cwebsocket.Close(conn, "", cwebsocket.STATUS_CLOSE_NORMAL_CLOSURE)
					return
				}
			}
		}()
	}
}

func netExample() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, _ := server.Accept()

		r, err := chttp.NewRequest(conn)
		if err != nil {
			slog.Info("failed parsing request", "ERR", err)
		}

		fmt.Println(r)

		chttp.NewTextResponse("Hello, World").Write(conn)

		conn.Close()
	}
}

// func nbioExample() {
// 	engine := nbio.NewEngine(nbio.Config{
// 		Network:            "tcp",
// 		Addrs:              []string{":8080"},
// 		MaxWriteBufferSize: 6 * 1024 * 1024,
// 	})

// 	engine.OnOpen(func(c *nbio.Conn) {
// 		log.Println("OnOpen:", c.RemoteAddr().String())
// 	})
// 	// hanlde connection closed
// 	engine.OnClose(func(c *nbio.Conn, err error) {
// 		log.Println("OnClose:", c.RemoteAddr().String(), err)
// 	})
// 	// handle data
// 	engine.OnData(func(c *nbio.Conn, data []byte) {
// 		r, err := chttp.NewRequest(c)
// 		if err != nil {
// 			slog.Info("failed parsing request", "ERR", err)
// 		}

// 		fmt.Println(r)

// 		chttp.NewTextResponse("Hello, World").Write(c)
// 	})

// 	err := engine.Start()
// 	if err != nil {
// 		log.Fatalf("nbio.Start failed: %v\n", err)
// 		return
// 	}
// 	defer engine.Stop()

// 	<-make(chan int)
// }

// func gnetExample() {
// 	var port int
// 	var multicore bool

// 	// Example command: go run echo.go --port 9000 --multicore=true
// 	flag.IntVar(&port, "port", 9000, "--port 8080")
// 	flag.BoolVar(&multicore, "multicore", false, "--multicore true")
// 	flag.Parse()
// 	echo := &echoServer{addr: fmt.Sprintf("tcp://:%d", port), multicore: multicore}
// 	log.Fatal(gnet.Run(echo, echo.addr, gnet.WithMulticore(multicore)))
// }

// type echoServer struct {
// 	gnet.BuiltinEventEngine

// 	eng       gnet.Engine
// 	addr      string
// 	multicore bool
// }

// func (es *echoServer) OnBoot(eng gnet.Engine) gnet.Action {
// 	es.eng = eng
// 	log.Printf("echo server with multi-core=%t is listening on %s\n", es.multicore, es.addr)
// 	return gnet.None
// }

// func (es *echoServer) OnTraffic(c gnet.Conn) gnet.Action {
// 	// buf, _ := c.Next(-1)
// 	// c.Write(buf)

// 	r, err := chttp.NewRequest(c)
// 	if err != nil {
// 		slog.Info("failed parsing request", "ERR", err)
// 	}

// 	fmt.Println(r)

// 	chttp.NewTextResponse("Hello, World").Write(c)

// 	return gnet.None
// }
