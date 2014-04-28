package main

import (
	"os"
	"log"
	"encoding/json"
	"net/http"
	"database/sql"
	"bytes"
	"os/exec"
	"fmt"
	"regexp"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
)

var (
	RegisterChan chan RegisterForm
	NameRegex = regexp.MustCompile("[\\w-]+")
	DefaultKeys = map[string]string {
		"akusete-root.pub":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC1GAwxTVwISOKpe6B95S2aq8/7vzQqrTol0+x7CQnpJmGxPhAUVDo/prSk3LjLNyRxp9RKpRXTrycLxQfHpr2uPi5gdGQO4eLSqaZJLB7Hpntv1shky8UFRag0C1AAFgtSNC5+onRyL5lHMFDd4nKwapkyPS2yA0iufquUynYUmRrIvFcY/NUmVYXXN+GhNtSdVAsgoWc1vKeL/xYBPEDG5lJIow5CExtXxc81yuuIS93EOb5V+hhHMl0Mi51BkG29xPjec9r4xOKsrspyLH5dpjQz7gf40K1ULG4BKeFaXTJ/u2EPugdmC2YMxE8RldTFjh0X2npSQ29xN2iNCH99 akusete-root",
		"admin":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBeuEqCv0wGYuroicaeAfAT6z3/Sf5H93VkkBK79n1X+ewaphkMpSA9ip5MvPEskniBYJWvpqNGWemIhiuu/ctqJVDGresffm9j/NoZ5O/z5P4wlm3psL+eavrP6HGs9MexkmcroJbUTN/B4hYC/EULuDtYEwkkFB4u1cJC6TdG616BA3Ilr8UKVYMWwsWUqFt4gyU5GY5M6Q5LDfNuZ6Q3dQ5/xFGLu8+GDcHrH7Jc8EhNCzaIquPuIPNr+P9wCQasmgj+BhNbfIE8PPP7jGIbQ2PgO4z2Sz+BAeuyCDOjh8OaE87rkIIUUsRzE7h9iDSQyAURpBdejECrIR13yOR webserver@ip-172-31-29-167",
	}
)

const (
	DbFile = "./users.db"
	Boilerplate = `
		[gitosis]

		[group gitosis-admin]
		writable = gitosis-admin
		members = admin akusete-root

	`
)

func main() {
	err := initDB()
	if err != nil { log.Fatal(err) }

	RegisterChan = make(chan RegisterForm)
	go registerMonitor()

	http.HandleFunc("/register", register)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func initDB() error {
	os.Remove(DbFile)

	db, err := sql.Open("sqlite3", DbFile)
	if err != nil { return err }
	defer db.Close()

	sql := "create table user (name text not null primary key, email text not null unique, public_key text not null);"
	_, err = db.Exec(sql)
	return err
}

func registerMonitor() {
	db, err := sql.Open("sqlite3", DbFile)
	if err != nil { log.Fatal(err) }
	defer db.Close()

	for form := range RegisterChan {
		err := addUser(db, form)
		if err != nil { log.Println(err) }
	}
}

func addUser(db *sql.DB, form RegisterForm) error {
	q := "insert into user(name, email, public_key) values(?,?,?)"
	stmt, err := db.Prepare(q)
	if err != nil {
		log.Printf("%v: %s\n", err, q)
		return err
	}
	_, err = stmt.Exec(form.Name, form.Email, form.PublicKey)
	if err != nil {
		log.Printf("%v: %s: %v\n", err, q, form)
		return err
	}
	log.Printf("added user %v\n", form)
	err = generateGitosis(db)
	return err
}



func generateGitosis(db *sql.DB) error {
	err := exec.Command("checkoutGitosis").Run()
	if err != nil { return err }

	var b bytes.Buffer
	b.WriteString(Boilerplate)

	rows, err := db.Query("select * from user")
	if err != nil { return err }

	for name, publicKey := range DefaultKeys {
	    err = ioutil.WriteFile(fmt.Sprintf("gitosis-admin/keydir/%s.pub", name), []byte(publicKey), 0755)
	    if err != nil { return err }
	}

	for rows.Next() {
	    var name string
	    var email string
	    var publicKey string
	    err = rows.Scan(&name, &email, &publicKey)
	    if err != nil { return err }

	    log.Println(name, email, publicKey)
	    b.WriteString(fmt.Sprintf("[group %s]\nwriteable = %s\nmembers = admin %s\n\n", name, name, name))
	    err = ioutil.WriteFile(fmt.Sprintf("gitosis-admin/keydir/%s.pub",name), []byte(publicKey), 0755)
	    if err != nil { return err }
	}
	err = ioutil.WriteFile("gitosis-admin/gitosis.conf", b.Bytes(), 0755)
	if err != nil { return err }

	err = exec.Command("updateGitosis").Run()
	return err
}

type RegisterForm struct {
	Name string `json:"name"`
	Email string `json:"email"`
	PublicKey string `json:"public_key"`
}

func (f *RegisterForm) IsValid() bool {
	notEmpty := f.Name != "" && f.Email != "" && f.PublicKey != ""
	_, reserved := DefaultKeys[f.Name]
	validName := NameRegex.MatchString(f.Name)
	return notEmpty && !reserved && validName
}

func register(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    var form RegisterForm
    err := decoder.Decode(&form)
    if err != nil {
    	form.Name = r.FormValue("name")
    	form.Email = r.FormValue("email")
    	form.PublicKey = r.FormValue("public_key")
    }
    if (form.IsValid()) {
		RegisterChan <- form
		w.WriteHeader(http.StatusOK)    	
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}