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


func email_check(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	var user User
	bs, _ := ioutil.ReadAll(request.Body)
	email := string(bs)

	q := datastore.NewQuery("Users").Filter("Email =", email)
	i, _ := q.Count(ctx)

	if i != 0{
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
	input := strings.Split(temp, "|")

	/* datastore NewKey */
	/*
	func NewKey(ctx context.Context, kind, name string, id int64, parent *Key) *Key
		NewKey creates a new key. kind cannot be empty. At least one of name and id must be zero. 
		If both are zero, the key returned is incomplete. parent must either be a complete key or nil.
	*/

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

	/*if err != nil && input[1] == input[2] || input[1] == "" || input[2] == ""{
		io.WriteString(response, "false")
	} else {
		io.WriteString(response, "true")
	}*/
}