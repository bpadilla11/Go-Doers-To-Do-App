package GodoersToDo


import (
	"io"
	"net/http"
	"io/ioutil"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"strings"
	"encoding/json"
)


//email check is used when creating a new user and updating user info
func email_check(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	var user User
	bs, _ := ioutil.ReadAll(request.Body)
	email := string(bs)

	//perform a datastore query with the given email received by ajax
	q := datastore.NewQuery("Users").Filter("Email =", email)
	i, _ := q.Count(ctx)

	//if the returned query contains 1 or more results then email
	//is not unique
	if i != 0{
		//below if statement is a guard when the user is updating his/her information.
		//get the user info from memcache, then compare the user email with the
		//email we got via ajax. if it is the same then it means the user did not change
		//his/her email.
		//we did this because if the user did not change email then this email
		//check will always say the email is not unique even though the email
		//is the current email of the user.
		item, _, err := getSession(request)
		json.Unmarshal(item.Value, &user)
		if err == nil{
			if user.Email == email{
				io.WriteString(response, "false")
				return
			}
		}

		io.WriteString(response, "true")
		return
	}
	io.WriteString(response, "false")
	return
}


func passw_check(response http.ResponseWriter, request *http.Request) {
	bs, _ := ioutil.ReadAll(request.Body)
	temp := string(bs)
	//string we receive via ajax is in the form password1|password2 so
	//we need to split it so we can compare.
	input := strings.Split(temp, "|")


	if input[0] == "" || input[1] == "" {
		io.WriteString(response, "false")
		return
	}
	if input[0] != input[1]{
		io.WriteString(response, "true")
		return
	}
	if input[0] == input[1]{
		io.WriteString(response, "false")
		return
	}
}


