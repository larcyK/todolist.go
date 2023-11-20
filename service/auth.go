package service

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"todolist.go/db"
	database "todolist.go/db"
)

func LoginCheck(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)
	if userID == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	}
	db, err := db.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		ctx.Abort()
	}
	var user database.User
	err = db.Get(&user, "SELECT * FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		ctx.Abort()
	}
	if !user.Valid {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func OwnershipCheck(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	db, err := db.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		ctx.Abort()
	} else {
		var ownership []database.Ownership
		err = db.Select(&ownership, "SELECT * FROM ownership WHERE task_id = ?", taskID)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			ctx.Abort()
		} else if len(ownership) == 0 {
			Error(http.StatusNotFound, "No such task")(ctx)
			ctx.Abort()
		} else if ownership[0].UserID != userID {
			Error(http.StatusForbidden, "You don't have permission to access this task")(ctx)
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}
