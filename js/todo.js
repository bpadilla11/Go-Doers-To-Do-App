var todo_content = document.querySelector("#todo-content");
var todo_image = document.querySelector("#todo-image");
var todo_list_all = document.querySelector("#todo-list-all");
var todo_list_queued = document.querySelector("#todo-list-queued");
var todo_list_done = document.querySelector("#todo-list-done");
var todo_submit = document.querySelector("#todo-submit");
var todo_form = document.querySelector("#todo-form");

var all = document.querySelector("#all");
var queued = document.querySelector("#queued");
var done = document.querySelector("#done");

var Todos = [];


//get the todos objects from the server
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

//create html element needed to render todo objects like this.
/*
<div id="todo-list">
	<div class="todo-style">
		<a href="#" class="delete" id="...">X</a>
		<p>1212</p><p class="time">Mon May 9 2016 04:29:00 PM</p>
	</div>
	<div class="todo-style">
		<a href="#" class="delete" id="...">X</a>
		<p>12</p><p class="time">Mon May 9 2016 04:29:00 PM</p>
	</div>
</div>
*/
function showTodos(){
	todo_list_all.innerHTML = "";
	todo_list_queued.innerHTML = "";
	todo_list_done.innerHTML = "";
	for(var i = 0; i < Todos.length; i++) {	
		todo_list_all.appendChild(createTodo(Todos[i]));
		if(Todos[i].Status == "queued"){
			todo_list_queued.appendChild(createTodo(Todos[i]));
			
		}
		else if(Todos[i].Status == "done"){
			todo_list_done.appendChild(createTodo(Todos[i]));
		}
	}
}

function createTodo(Todos){
		var p = document.createElement("p");
		var div = document.createElement("div");
		var img = document.createElement("img");
		var img_a = document.createElement("a");
		var a_delete = document.createElement("button");
		var time  = document.createElement("p");
		var status = document.createElement("button");

		if(Todos.Status == "done")
			status.className = "fa fa-check-square status status-done";
		else
			status.className = "fa fa-check-square status";
		status.id = Todos.ToDoId;
		status.addEventListener('click', function (e) {
			var id = e.target.id;
			var xhr = new XMLHttpRequest();
        	xhr.open("UPDATE", "/todo?todo="+id);
        	xhr.send(null);
        	xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                setTimeout(getTodos, 100);
            }
        }
		});

		a_delete.className = "delete";
		a_delete.innerHTML = "X";
		a_delete.id = Todos.ToDoId;
		a_delete.addEventListener('click', function (e) {
			var id = e.target.id;
			var xhr = new XMLHttpRequest();
        	xhr.open("DELETE", "/todo?todo="+id);
        	xhr.send(null);
        	xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
                setTimeout(getTodos, 100);
            }
        }
		});

		img_a.setAttribute("href", Todos.Photo_Link);
		img_a.setAttribute("target", "_blank");

		img.setAttribute("src", Todos.Photo_Media);
		img.className = "todo-photo";
		img_a.appendChild(img);


		p.innerHTML = Todos.Content;

		time.innerHTML = Todos.Date;
		time.className = "time";

		div.appendChild(status);
		div.appendChild(a_delete);
		div.appendChild(p);
		if(Todos.Photo_Media != "")
			div.appendChild(img_a);
		div.appendChild(time);
		div.className = "todo-style";
		return div
}

//to add a todo object
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


all.addEventListener('click', function (e) {
	all.className = "active";
	queued.className = "";
	done.className = "";
    showTodos();
    todo_list_all.className = "todo-list";
    todo_list_queued.className = "";
    todo_list_done.className = "";
});

queued.addEventListener('click', function (e) {
	all.className = "";
	queued.className = "active";
	done.className = "";
	showTodos();
    todo_list_all.className = "";
    todo_list_queued.className = "todo-list";
    todo_list_done.className = "";
});

done.addEventListener('click', function (e) {
	all.className = "";
	queued.className = "";
	done.className = "active";
	showTodos();
    todo_list_all.className = "";
    todo_list_queued.className = "";
    todo_list_done.className = "todo-list";
});