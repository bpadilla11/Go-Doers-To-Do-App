var email = document.querySelector("#email");
var password1 = document.querySelector("#password1");
var password2 = document.querySelector("#password2")


//check unique email
email.addEventListener('input', function(){
	var xhr = new XMLHttpRequest();
	xhr.open('POST', '/api/email_check');
	xhr.send(email.value);
	xhr.addEventListener('readystatechange', function(){
		if (xhr.readyState === 4 && xhr.status === 200){
			var taken = xhr.responseText;
			if (taken == 'true'){
				//not unique email
			}
			else{
				//unique email
			}
		}
	});
});


//password match
password2.addEventListener('input', function(){
	if(password1 != password2) {
		//password not match
	}
});