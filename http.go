package GodoersToDo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handleGetGif)
}

func main(){

}

func handleGetGif(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	t := req.FormValue("term")
	if t == "" {
		t = "rocket+league"
	}


	client := urlfetch.Client(ctx)
	result, err := client.Get("http://api.giphy.com/v1/gifs/search?q=" + t + "&api_key=dc6zaTOxFJmzC")
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	defer result.Body.Close()

	var obj struct {
		Data []struct {
			URL string ""
			Images struct {
				Original struct {
					URL string
				}
			}
		}
	}
	err = json.NewDecoder(result.Body).Decode(&obj)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	img := obj.Data[1] 
	fmt.Fprintf(res, `<a href="%v"></a><img src="%v"><br>`, img.URL, img.Images.Original.URL)
}
