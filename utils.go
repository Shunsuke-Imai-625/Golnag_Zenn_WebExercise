package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

//クッキーにセッションがあればユーザー情報を取得する
func getUser(w http.ResponseWriter, req *http.Request) user {
	var u user
	//クッキーを取得
	c, err := req.Cookie("session")
	if err != nil {

		sID := uuid.New()
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

	}
	c.MaxAge = sessionLength
	http.SetCookie(w, c)

	//セッションがあればユーザーDB情報を取得
	if sess, ok := dbSessions[c.Value]; ok {
		sess.ltime = time.Now()
		dbSessions[c.Value] = sess
		u = dbUsers[sess.un]
	}
	return u
}

//ログイン状態かどうかをチェックする
func alreadyLoggedIn(w http.ResponseWriter, req *http.Request) bool {
	//クッキーを取得
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}
	c.MaxAge = sessionLength
	http.SetCookie(w, c)

	//ユーザーDB情報を取得
	sess, ok := dbSessions[c.Value]
	if ok {
		sess.ltime = time.Now()
		dbSessions[c.Value] = sess
	}
	_, ok = dbUsers[sess.un]
	return ok

}

func showSessions() {
	fmt.Println("***action=showSessions()")
	for k, v := range dbSessions {
		fmt.Println(k, v.un, v.ltime)
	}
	fmt.Println("")
}

func cleansession() {
	fmt.Println("***before_action=cleanSession()")
	for k, v := range dbSessions {
		fmt.Println(k, v)
		if time.Now().Sub(v.ltime) > (time.Second * 30) {
			delete(dbSessions, k)
		}

	}
	fmt.Println("***before_action=cleanSession()")
	for k, v := range dbSessions {
		fmt.Println(k, v)
	}
	fmt.Println("")
}
