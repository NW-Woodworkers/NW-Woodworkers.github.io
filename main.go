package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nwwa/NW-Woodworkers.github.io/sheets"
)

var (
	masterTmpl *template.Template

	// spreadsheetID is the Google Sheets document which will be appended to
	// with new user information
	spreadsheetID = "1WcQYPqi_8OaUsTLjb2k51rAErC3Hn66qcoDNmDLbCnM"

	sheetsWriter sheets.Appender
)

func logError(r *http.Request, err error, msg string) {
	// TODO: Do something with the request to give more context
	// Such as printing the email, phone, and name so I can contact them
	fmt.Printf("%s: %s\n", msg, err)
}

func renderSignupFailure(w http.ResponseWriter) {
	tmpl, _ := template.Must(masterTmpl.Clone()).ParseFiles("templates/signup-failure.html")
	err := tmpl.Execute(w, map[string]string{"title": "Membership Failed"})
	if err != nil {
		logError(nil, err, "Failed to render the signup-failure page")
	}
}

func handleSignupForm(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.Must(masterTmpl.Clone()).ParseFiles("templates/signup.html")
	err := tmpl.Execute(w, map[string]string{"title": "Membership Signup"})
	if err != nil {
		logError(r, err, "Failed to render the signup page")
	}
}

func handlePaymentForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// TODO: Handle this error
	}

	email := r.Form.Get("email")
	if email == "" || !strings.Contains(email, "@") {
		// TODO: Error that the email is required
	}

	fullNameParts := strings.Split(r.Form.Get("name"), " ")
	if len(fullNameParts) < 2 {
		// TODO: Error that there must be a first and last name!
		// This should just be filled in to the previous template
	}
	firstName := strings.Join(fullNameParts[:len(fullNameParts)-1], " ")

	info := sheets.SignupInfo{
		FirstName: firstName,
		LastName:  fullNameParts[len(fullNameParts)-1],
		Addr:      r.Form.Get("address"),
		City:      r.Form.Get("city"),
		State:     r.Form.Get("state"),
		Zip:       r.Form.Get("zipcode"),
		Phone:     r.Form.Get("phone"),
		Email:     email,
		Since:     time.Now().Format("Jan 2 2006"),
	}

	err = sheetsWriter.WriteNewSignup(info)
	if err != nil {
		logError(r, err, "Failed to add user to membership spreadsheet")
		renderSignupFailure(w)
		return
	}

	tmpl, _ := template.Must(masterTmpl.Clone()).ParseFiles("templates/payment.html")
	err = tmpl.Execute(w, map[string]string{
		"title": "Membership Signup",
		"email": email,
	})
	if err != nil {
		logError(r, err, "Failed render the payment page after signup")
		renderSignupFailure(w)
		return
	}
}

func main() {
	masterTmpl, _ = template.ParseFiles("templates/master.html")
	http.HandleFunc("/signup", handleSignupForm)
	http.HandleFunc("/payment", handlePaymentForm)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	var err error
	sheetsWriter, err = sheets.New(spreadsheetID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening...")
	err = http.ListenAndServe(":8586", nil)
	panic(err.Error())
}
