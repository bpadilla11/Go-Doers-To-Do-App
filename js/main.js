function show_dropdown_form(){
	var form = document.querySelector("#dropdown-form");
	if(form.style.display == "none"){
		form.style.display = "flex";
	}
	else{
		form.style.display = "none";
	}
}


function show_login_modal(status){
	var modal = document.querySelector("#login_status_modal");
	if(status != ""){
		modal.className = "modalDialog_show";
	}
	else{
		modal.className = "modalDialog";
	}
}

function close_login_modal(){
	document.querySelector("#login_status_modal").className = "modalDialog";
}