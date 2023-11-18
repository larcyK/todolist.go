package service

import (
	"crypto/sha256"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_user.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func isValidPassword(password string) bool {
	if utf8.RuneCountInString(password) < 8 {
		return false
	}

	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	return hasDigit && hasLetter
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	passwordConfirm := ctx.PostForm("password_confirm")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "form_new_user.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
		return
	case password == "":
		ctx.HTML(http.StatusBadRequest, "form_new_user.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの確認
	if password != passwordConfirm {
		ctx.HTML(http.StatusBadRequest, "form_new_user.html", gin.H{"Title": "Register user", "Error": "Password does not match", "Username": username, "Password": password})
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "form_new_user.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}

	// パスワードの複雑さの確認 8文字以上かつ英数字を含む
	if !isValidPassword(password) {
		ctx.HTML(http.StatusBadRequest, "form_new_user.html", gin.H{"Title": "Register user", "Error": "Password is too simple", "Username": username, "Password": password})
		return
	}

	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user)
}
