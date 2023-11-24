package service

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	dn := ctx.Query("dn")
	sort := ctx.Query("sort")

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage := 5
	offset := (page - 1) * perPage

	// Get tasks in DB
	var tasks []database.Task

	var conditions []string
	if kw != "" {
		conditions = append(conditions, fmt.Sprintf("(title LIKE '%%%s%%' OR detail LIKE '%%%s%%')", kw, kw))
	}
	switch dn {
	case "done":
		conditions = append(conditions, fmt.Sprintf("is_done=%t", true))
	case "not_done":
		conditions = append(conditions, fmt.Sprintf("is_done=%t", false))
	}
	var sortQuery string = ""
	switch sort {
	case "deadline":
		sortQuery = "ORDER BY deadline"
	case "id":
		sortQuery = "ORDER BY id"
	case "title":
		sortQuery = "ORDER BY title"
	default:
		sortQuery = "ORDER BY id"
	}

	conditions = append(conditions, fmt.Sprintf("user_id=%d", userID))

	query := "SELECT tasks.* FROM tasks INNER JOIN ownership ON task_id = id "
	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, c := range conditions[1:] {
			query += " AND " + c
		}
	}
	query += " " + sortQuery

	// query += fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	err = db.Select(&tasks, query)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	startTask := min(len(tasks), offset)
	endTask := min(len(tasks), startTask+perPage)
	TotalPage := (len(tasks)-1)/perPage + 1

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{
		"Title":     "Task list",
		"Tasks":     tasks[startTask:endTask],
		"Kw":        kw,
		"Dn":        dn,
		"Sort":      sort,
		"User":      userID,
		"Page":      page,
		"TotalPage": TotalPage,
	})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get ownerships in given task
	var userNames []string
	err = db.Select(&userNames, "SELECT name FROM users INNER JOIN ownership ON user_id = id WHERE task_id=? AND valid=1", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render task
	// ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task_info.html", gin.H{"Title": task.Title, "Task": task, "Owners": userNames})
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	if title == "" {
		ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration", "Error": "Title is empty"})
		return
	}
	// Get task detail
	detail, exist := ctx.GetPostForm("detail")
	if !exist {
		Error(http.StatusBadRequest, "No detail is given")(ctx)
		return
	}
	// Get task deadline
	clientLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	deadline, exist := ctx.GetPostForm("deadline")
	// fmt.Printf("deadline: %s\n", deadline)
	if !exist {
		Error(http.StatusBadRequest, "No deadline is given")(ctx)
		return
	}
	deadlineTime, err := time.ParseInLocation("2006-01-02T15:04", deadline, clientLocation)
	if err != nil {
		ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration", "Error": "Invalid deadline"})
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Register task
	tx := db.MustBegin()
	// Create new data with given title on DB
	result, err := tx.Exec("INSERT INTO tasks (title, detail, deadline) VALUES (?, ?, ?)", title, detail, deadlineTime)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context) {
	// Get task ID
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	// Get task detail
	detail, exist := ctx.GetPostForm("detail")
	if !exist {
		Error(http.StatusBadRequest, "No detail is given")(ctx)
		return
	}
	// Get task deadline
	clientLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
		Error(http.StatusBadRequest, "No deadline is given")(ctx)
		return
	}
	deadlineTime, err := time.ParseInLocation("2006-01-02T15:04", deadline, clientLocation)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get task Status
	isDone_s, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No status is given")(ctx)
		return
	}
	isDone, err := strconv.ParseBool(isDone_s)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Update task
	tx := db.MustBegin()
	_, err = tx.Exec("UPDATE tasks SET title=?, detail=?, deadline=?, is_done=? WHERE id=?", title, detail, deadlineTime, isDone, id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	// Render status
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx := db.MustBegin()
	// Delete the task from DB
	_, err = tx.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("DELETE FROM ownership WHERE task_id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	// Redirect to /list
	// ctx.Redirect(http.StatusFound, "/list")
}

func ShareTaskForm(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_share_task.html", gin.H{"Title": fmt.Sprintf("Share task %d", id), "Task": task})
}

func ShareTask(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get user ID
	shareUserID, exist := ctx.GetPostForm("user_id")
	if !exist {
		Error(http.StatusBadRequest, "No user ID is given")(ctx)
		return
	}
	myUserID := sessions.Default(ctx).Get("user")
	if myUserID == shareUserID {
		Error(http.StatusBadRequest, "You can't share your task to yourself")(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Check if the user exists
	var user database.User
	err = db.Get(&user, "SELECT * FROM users WHERE id=?", shareUserID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Register task
	_, err = db.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", shareUserID, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render status
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}
