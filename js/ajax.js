var email = document.querySelector("#register-email");
var password1 = document.querySelector("#register-password1");
var password2 = document.querySelector("#register-password2")
var register_ajax_error1 = document.querySelector("#register-ajax-error1");
var register_ajax_error2 = document.querySelector("#register-ajax-error2");


//check unique email
email.addEventListener('input', function(){
	var xhr = new XMLHttpRequest();
	xhr.open('POST', '/api/email_check');
	xhr.send(email.value);
	xhr.addEventListener('readystatechange', function(){
		if (xhr.readyState === 4 && xhr.status === 200){
			var err = xhr.responseText;
			if (err == 'true'){
				register_ajax_error1.innerHTML = "* Email Taken!"
				register_ajax_error1.style.display = "flex";
			}
			else{
				register_ajax_error1.style.display = "none";
				register_ajax_error1.innerHTML = "";
			}
		}
	});
});


//check password confirmation
//one for password1 and password2
password1.addEventListener('input', function(){
	var xhr = new XMLHttpRequest();
	xhr.open('POST', '/api/passw_check');
	xhr.send(password1.value + "|" + password2.value);
	xhr.addEventListener('readystatechange', function(){
		if (xhr.readyState === 4 && xhr.status === 200){
			var err = xhr.responseText;
			if (err == 'true'){
				register_ajax_error2.innerHTML = "* Passwords not match!"
				register_ajax_error2.style.display = "flex";
			}
			else{
				register_ajax_error2.style.display = "none";
				register_ajax_error2.innerHTML = "";
			}
		}
	});
});

password2.addEventListener('input', function(){
	var xhr = new XMLHttpRequest();
	xhr.open('POST', '/api/passw_check');
	xhr.send(password1.value + "|" + password2.value);
	xhr.addEventListener('readystatechange', function(){
		if (xhr.readyState === 4 && xhr.status === 200){
			var err = xhr.responseText;
			if (err == 'true'){
				register_ajax_error2.innerHTML = "* Passwords not match!"
				register_ajax_error2.style.display = "flex";
			}
			else{
				register_ajax_error2.style.display = "none";
				register_ajax_error2.innerHTML = "";
			}
		}
	});
});
