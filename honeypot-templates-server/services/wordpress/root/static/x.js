function report() {
    var username = $("#user_login").val();
    var password = $("#user_pass").val();

    $.ajax({
        type: "POST",
        url: "/login",
        dataType: "json",
        data: {
            "username": username,
            "password": password,
        },
        success: function (e) {
            if (e.code == 0) {
                alert("用户名或密码错误");
            } else {
                console.log(e.msg)
            }
        },
        error: function (e) {
            console.log("fail")
        }
    });
}