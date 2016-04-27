package GodoersToDo


import (
	"io"
	"net/http"
	"io/ioutil"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"strings"
)


func email_check(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	bs, err := ioutil.ReadAll(request.Body)
	var user User
	temp := string(bs)
	//input := strings.Split(temp, "|")
	user.Email = temp

	/* datastore NewKey */
	/*
	func NewKey(ctx context.Context, kind, name string, id int64, parent *Key) *Key
		NewKey creates a new key. kind cannot be empty. At least one of name and id must be zero. 
		If both are zero, the key returned is incomplete. parent must either be a complete key or nil.
	*/
	key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
	err = datastore.Get(ctx, key, &user)
	if err == nil {
		io.WriteString(response, "true")
		return
	}
	io.WriteString(response, "false")
	return

	/*if err != nil && input[1] == input[2] || input[1] == "" || input[2] == ""{
		io.WriteString(response, "false")
	} else {
		io.WriteString(response, "true")
	}*/
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