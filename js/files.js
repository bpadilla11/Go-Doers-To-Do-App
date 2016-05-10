var file_list = document.querySelector("#file-list");
var Files = [];

//get the files from the gcs for the specific user
function getFiles() {
	var xhr = new XMLHttpRequest();
        xhr.open("GET", "/api/filehelper");
        xhr.send(null);
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                Files = JSON.parse(xhr.responseText); //list of files
                showFiles();
            }
        }
}

getFiles();

//show the files from the gcs for the specific user
/*
<div id="file-list">
	<figure>
		<a href="..." target="_blank">
			<img src="...">
		</a>
		<figcaption>
			<a href="..." class="link">Download</a>
			<a href="..." target="_blank" class="name">filename</a>
			<a class="delete-file" href="#" id="...">Delete</a>
		</figcaption>
	</figure>
</div>
*/
function showFiles(){
	file_list.innerHTML = "";
	if(Files != null)
		for(var i = 0; i < Files.length; i++) {
			figure = document.createElement("figure");
			view = document.createElement("a");
			img = document.createElement("img");
			figcaption = document.createElement("figcaption");
			download = document.createElement("a");
			filename = document.createElement("a");
			delete_file = document.createElement("a");

			view.setAttribute("href", Files[i].Source_Link);
			view.setAttribute("target", "_blank");
			img.setAttribute("src", Files[i].Download_Link);
			img.className = "photo";
			view.appendChild(img);

			download.setAttribute("href", Files[i].Download_Link);
			download.className = "link";
			download.innerHTML = "Download";

			filename.innerHTML = Files[i].Name;
			filename.setAttribute("href", Files[i].Source_Link);
			filename.setAttribute("target", "_blank");
			filename.className = "name";

			delete_file.className = "delete-file";
			delete_file.setAttribute("href", "#");
			delete_file.id = Files[i].Name;
			delete_file.innerHTML = "Delete"

			figcaption.appendChild(download);
			figcaption.appendChild(filename);
			figcaption.appendChild(delete_file);

			figure.appendChild(view)
			figure.appendChild(figcaption)
			file_list.appendChild(figure)
		}
		else{
			file_list.innerHTML = "No Files";
		}
}


//delete a file effectively deleting it from the gcs via our go server
(function () {
    file_list.addEventListener("click", function (evt) {
    	if(evt.target.className == "delete-file"){
        var id = evt.target.id;
		var xhr = new XMLHttpRequest();
        xhr.open("DELETE", "/api/filehelper?filename="+id);
        xhr.send(null);
      	var it = document.getElementById(id);
      	if(it != null){
      		it.innerHTML = "Deleting...";
      		it.className = "deleting";
      	}
        xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            setTimeout(getFiles, 100);
            }
        };  
        }              
    }, false);
})();