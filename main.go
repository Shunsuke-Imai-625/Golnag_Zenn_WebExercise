package main

import (
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var tpl *template.Template

type user struct {
	UserName   string
	Password   []byte // change to []byte
	First      string
	Last       string
	Permission string
}

type session struct {
	un    string
	ltime time.Time
}

//var dbSessions = map[string]string{} // session ID, user ID
var dbSessions = map[string]session{}
var dbUsers = map[string]user{} //userID, user info

const sessionLength int = 30

func init() {
	tpl = template.Must(template.ParseGlob("app/views/templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/vip", vip)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)

}

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	showSessions()
	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func vip(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	if !alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	if u.Permission != "dog" {
		http.Error(w, "No dogs allowed.", http.StatusForbidden)
		return
	}
	tpl.ExecuteTemplate(w, "vip.gohtml", u)
}

func signup(w http.ResponseWriter, req *http.Request) {

	//すでにログインしている場合はこのページは必要ない
	if alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	//FORMから送信してきた時の処理
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")
		s := req.FormValue("permission")

		//もしユーザーネームが既に使われていたらエラーとする
		if _, ok := dbUsers[un]; ok {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}

		//セッションDBを作成
		sID := uuid.New()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		c.MaxAge = sessionLength
		http.SetCookie(w, c)
		dbSessions[c.Value] = session{un, time.Now()}

		//パスワードを暗号化して保存
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		u := user{un, bs, f, l, s}
		dbUsers[un] = u

		//リダイレクト
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

func login(w http.ResponseWriter, req *http.Request) {

	//ログイン済みの場合はこのページが不要
	if alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	//情報を受け取ったときの処理
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		//ユーザーDBから情報を持ってくる（なかったらエラー）
		u, ok := dbUsers[un]
		if !ok {
			http.Error(w, "USER NAME - PASSWORD not match", http.StatusForbidden)
			return
		}
		//受信したパスワードとDBのパスワードの照合(異なっていたらエラー)
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "USER NAME - PASSWORD not match", http.StatusForbidden)
			return
		}

		//セッションを作成する
		sID := uuid.New()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		c.MaxAge = sessionLength
		http.SetCookie(w, c)
		dbSessions[c.Value] = session{un, time.Now()}
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}

func logout(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther) //303
		return
	}
	c, _ := req.Cookie("session")
	//セッションから抜ける
	delete(dbSessions, c.Value)
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	cleansession()
	http.Redirect(w, req, "/login", http.StatusSeeOther) //303
}
