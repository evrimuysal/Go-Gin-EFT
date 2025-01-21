package main

import (
	"gin-mongo-api/configs"
	"gin-mongo-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
)

const ErrPortNotSet = "$PORT must be set"

var (
	clients     = make(map[*websocket.Conn]string)
	rooms       = make(map[string]map[*websocket.Conn]bool)
	clientsLock = sync.Mutex{}
	roomsLock   = sync.Mutex{}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		log.Fatal(ErrPortNotSet)
	}

	router := setupRouter()
	router.Run(":" + portEnv)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ws", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to upgrade to WebSocket:", err)
			return
		}
		defer conn.Close()

		callerId := c.Query("callerId")
		if callerId == "" {
			log.Println("callerId missing in query")
			return
		}
		log.Println(callerId, "Connected")

		registerClient(conn, callerId)
		defer unregisterClient(conn, callerId)

		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}
			handleMessage(conn, callerId, message)
		}
	})
	configs.ConnectDB()

	router.Use(gin.Logger())
	socket := router.Group("/")
	socket.Use(callerIDMiddleware())
	routes.UserRoute(router)
	router.SetTrustedProxies([]string{})
	return router
}
func registerClient(conn *websocket.Conn, callerId string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	clients[conn] = callerId

	joinRoom(callerId, conn)
}

func unregisterClient(conn *websocket.Conn, callerId string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	delete(clients, conn)

	leaveRoom(callerId, conn)
}

func joinRoom(roomId string, conn *websocket.Conn) {
	roomsLock.Lock()
	defer roomsLock.Unlock()
	if rooms[roomId] == nil {
		rooms[roomId] = make(map[*websocket.Conn]bool)
	}
	rooms[roomId][conn] = true
}

func leaveRoom(roomId string, conn *websocket.Conn) {
	roomsLock.Lock()
	defer roomsLock.Unlock()
	if rooms[roomId] != nil {
		delete(rooms[roomId], conn)
		if len(rooms[roomId]) == 0 {
			delete(rooms, roomId)
		}
	}
}

func handleMessage(conn *websocket.Conn, senderId string, message map[string]interface{}) {
	event, ok := message["event"].(string)
	if !ok {
		log.Println("Invalid message format: missing event")
		return
	}

	switch event {
	case "ping":
		err := conn.WriteJSON(map[string]interface{}{
			"event": "pong",
			"data":  message["data"],
		})
		if err != nil {
			log.Println("Error sending pong:")
		}

	case "call":
		calleeId := message["calleeId"].(string)
		rtcMessage := message["rtcMessage"]
		forwardMessageToRoom(calleeId, "newCall", map[string]interface{}{
			"callerId":   senderId,
			"rtcMessage": rtcMessage,
		})

	case "answerCall":
		callerId := message["callerId"].(string)
		rtcMessage := message["rtcMessage"]
		forwardMessageToRoom(callerId, "callAnswered", map[string]interface{}{
			"callee":     senderId,
			"rtcMessage": rtcMessage,
		})

	case "callEnding":
		callerId := message["callerId"].(string)
		forwardMessageToRoom(callerId, "callEnd", map[string]interface{}{})

	case "ICEcandidate":
		calleeId := message["calleeId"].(string)
		rtcMessage := message["rtcMessage"]
		forwardMessageToRoom(calleeId, "ICEcandidate", map[string]interface{}{
			"sender":     senderId,
			"rtcMessage": rtcMessage,
		})

	default:
		log.Println("Unknown event:", event)
	}
}

func forwardMessageToRoom(roomId, event string, data map[string]interface{}) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	if connections, ok := rooms[roomId]; ok {
		for conn := range connections {
			err := conn.WriteJSON(map[string]interface{}{
				"event": event,
				"data":  data,
			})
			if err != nil {
				log.Println("Error forwarding message:", err)
			}
		}
	}
}
func callerIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		callerId := c.Query("callerId")
		if callerId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "callerId is required",
			})
			log.Println("Middleware blocked: callerId is missing in query")
			c.Abort()
			return
		}
		c.Next()
	}
}
