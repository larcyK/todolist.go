package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
	store := cookie.NewStore([]byte("my-secret"))
	engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)

	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		taskGroup.GET("/:id", service.OwnershipCheck, service.ShowTask) // ":id" is a parameter
		// タスクの新規登録
		taskGroup.GET("/new", service.NewTaskForm)
		taskGroup.POST("/new", service.RegisterTask)
		// 既存タスクの編集
		taskGroup.GET("/edit/:id", service.OwnershipCheck, service.EditTaskForm)
		taskGroup.POST("/edit/:id", service.OwnershipCheck, service.UpdateTask)
		// 既存タスクの削除
		taskGroup.DELETE("/delete/:id", service.OwnershipCheck, service.DeleteTask)
	}

	userGroup := engine.Group("/user")
	userGroup.Use(service.LoginCheck)
	{
		// ユーザ情報
		userGroup.GET("info", service.UserInfo)
		// ユーザ情報の編集
		userGroup.GET("/edit", service.EditUserForm)
		userGroup.POST("/edit", service.UpdateUser)
		// userGroup.POST("/edit", service.UpdateUser)
		// ユーザの削除
		userGroup.DELETE("/delete", service.DeleteUser)
	}

	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)
	// ログイン
	engine.GET("/login", service.LoginForm)
	engine.POST("/login", service.Login)
	// ログアウト
	engine.GET("/logout", service.Logout)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))

}
