var todo_content = document.querySelector("#todo-content");
var todo_image = document.querySelector("#todo-image");
var todo_list = document.querySelector("#todo-list");
var todo_submit = document.querySelector("#todo-submit");
var todo_form = document.querySelector("#todo-form");

var Todos = [];

function getTodos() {
	var xhr = new XMLHttpRequest();
        xhr.open("GET", "/todo?todo=");
        xhr.send(null);
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                Todos = JSON.parse(xhr.responseText);
                showTodos();
            }
        }
}

getTodos();

function showTodos(){
	todo_list.innerHTML = "";
	for(var i = 0; i < Todos.length; i++) {
		var p = document.createElement("p");
		var div = document.createElement("div");
		var img = document.createElement("img");
		var img_a = document.createElement("a");
		var a_delete = document.createElement("a");
		var time  = document.createElement("p");

		a_delete.setAttribute("href", "#");
		a_delete.className = "delete";
		a_delete.innerHTML = "X";
		a_delete.id = Todos[i].ToDoId;

		img_a.setAttribute("href", Todos[i].Photo_Link);
		img_a.setAttribute("target", "_blank");

		img.setAttribute("src", Todos[i].Photo_Media);
		img.className = "todo-photo";
		img_a.appendChild(img);


		p.innerHTML = Todos[i].Content;

		time.innerHTML = "Added: " + Todos[i].Date;
		time.className = "time";

		div.appendChild(a_delete);
		div.appendChild(p);
		if(Todos[i].Photo_Media != "")
			div.appendChild(img_a);
		div.appendChild(time);
		div.className = "todo-style";
		todo_list.appendChild(div);
	}
}


todo_submit.addEventListener('click', function (e) {
    var formData = new FormData();
    var content = todo_content.value;
	var file = todo_image.files[0];
	var xhr = new XMLHttpRequest();
	formData.append('content', content);
	formData.append('file', file);
	xhr.open('POST', '/todo');
	xhr.send(formData);
	todo_content.value = "";
	todo_image.value = null;
	todo_submit.className = "btn todo-submit-disable";
	todo_submit.innerHTML = "Please wait...";
	xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
    	    var item = xhr.responseText;
            item = JSON.parse(item);
            if(item.Photo_Media != "invalid")
            	Todos.push(item);	
            else
            	show_login_modal("invalid");
            todo_submit.className = "btn";
            todo_submit.innerHTML = "Add";
            showTodos();
        }
    };
});


(function () {
    todo_list.addEventListener("click", function (evt) {
        var id = evt.target.id;
        var xhr = new XMLHttpRequest();
		var xhr = new XMLHttpRequest();
        xhr.open("DELETE", "/todo?todo="+id);
        xhr.send(null);
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                setTimeout(getTodos, 100);
            }
        };          
    }, false);
})();