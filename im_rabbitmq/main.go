package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-chatery/im_rabbitmq/mq"
	"github.com/yzletter/go-chatery/im_rabbitmq/mysql"
	"github.com/yzletter/go-chatery/im_rabbitmq/server"
)

func ListenTermSignal(RabbitMQSvc *mq.RabbitMQService) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")
	RabbitMQSvc.Release()
	os.Exit(0)
}

func main() {

	engine := gin.Default()
	mysql.InitDB()
	RabbitMQSvc := mq.NewRabbitMQService()

	go ListenTermSignal(RabbitMQSvc)

	UserHdl := server.NewUserHandler(RabbitMQSvc)

	engine.GET("/register", UserHdl.Register)
	engine.GET("/speak/:id/:target", UserHdl.Speak)
	if err := engine.Run("localhost:8081"); err != nil {
		fmt.Println("服务启动失败")
		panic(err)
	}
}
