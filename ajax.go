package GodoersToDo


import (
	"io"
	"net/http"
	"io/ioutil"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)


func email_check(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	bs, err := ioutil.ReadAll(request.Body)
	var user User
	user.Email = string(bs)

	/* datastore NewKey */
	/*
	func NewKey(ctx context.Context, kind, name string, id int64, parent *Key) *Key
		NewKey creates a new key. kind cannot be empty. At least one of name and id must be zero. 
		If both are zero, the key returned is incomplete. parent must either be a complete key or nil.
	*/
	key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
	err = datastore.Get(ctx, key, &user)
	if err != nil{
		tpl.ExecuteTemplate(response, "register.html", nil)
		return
	}
	io.WriteString(response, "true")
}