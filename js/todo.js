var todo_content = document.querySelector("#todo-content");
var todo_image = document.querySelector("#todo-image");
var todo_list = document.querySelector("#todo-list");
var todo_submit = document.querySelector("#todo-submit");

var Todos = []

function getTodos() {
	var xhr = new XMLHttpRequest();
        xhr.open("GET", "/todo");
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
	//clear the todo_list
	todo_list.innerHTML = "";
	for(var i = 0; i < Todos.length; i++) {
		var h2 = document.createElement("h2");
		var div = document.createElement("div");
		var img = document.createElement("img");
		img.setAttribute("src", Todos[i].Photo);
		h2.innerHTML = Todos[i].Content;
		div.appendChild(h2);
		div.appendChild(img);
		todo_list.appendChild(div);
	}
}


// add new item
todo_submit.addEventListener('click', function (e) {
    var formData = new FormData();
    var content = document.querySelector('#todo-content').value;
	var file = document.querySelector('#todo-image').files[0];
	var xhr = new XMLHttpRequest();
	formData.append('content', content);
	formData.append('file', file);
	xhr.open('POST', '/todo');
	xhr.send(formData);
	xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
    	    var item = xhr.responseText;
            item = JSON.parse(item);
            Todos.push(item);
            showTodos();
        }
    };
});


/*
todo_submit.addEventListener('click', function(){
	var content = todo_content.value;
	todo_content.value = "";

	var formData = new FormData(), 
		file = document.querySelector("#todo-image").files[0],
		xhr = new XMLHttpRequest();
	formData.append('content', content)
	formData.append('file', file);
	xhr.open('POST', '/api/todo');
	xhr.send(formData);
}, false);*/
