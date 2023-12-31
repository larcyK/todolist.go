package main

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

func one_until(n int) []int {
	a := make([]int, n)
	for i := 0; i < n; i++ {
		a[i] = i + 1
	}
	return a
}

func add(a, b int) int {
	return a + b
}

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
	engine.SetFuncMap(template.FuncMap{
		"one_until": one_until,
		"add":       add,
	})

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
		// タスクの共有
		taskGroup.GET("/share/:id", service.OwnershipCheck, service.ShareTaskForm)
		taskGroup.POST("/share/:id", service.OwnershipCheck, service.ShareTask)
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
		// パスワードの変更
		userGroup.GET("/password", service.EditPasswordForm)
		userGroup.POST("/password", service.UpdatePassword)
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
