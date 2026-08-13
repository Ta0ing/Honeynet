var Login = (function () {
    var tip = $('.tip');
    var mask = $('.mask')[0];
    var captcha = $("#captcha");
    var name = $("#name");
    var pwd = $("#password");
    var is_checked = $('#is_checked');
    var language_select = $('#language_select');

    if (__permission == "") {
        document.getElementsByClassName('input_div_checkbox')[0].style.display = 'none';
        is_checked.value = 'on';
    } else {
        //许可协议内容
        if (is_checked.value == 'on') {
            is_checked.setAttribute('checked', 'true');
            $('#btn_submit').className += ' btn_hover';
        } else {
            is_checked.removeAttribute('checked');
        }
    }
    //表单提交
    function form_submit() {
        var username = name.value;
        var password = pwd.value;
        var verify = captcha.value;
        var language = language_select.value;
        clearShowMsg();
        if (username == "") {
            showMsg(lang.pleaseName, tip[0]);
            change_captcha();
            return;
        } else if (password == "") {
            showMsg(lang.pleasePW, tip[0]);
            change_captcha();
            return;
        } else if (captcha == '') {
            showMsg(lang.pleaseCAPTCHA, tip[1]);
            change_captcha();
            return;
        } else if (language == '') {
            showMsg(lang.infoerror, tip[0]);
            change_captcha();
            return;
        }
        var data = {
            name: crypt(username,'base64'),
            password: crypt(password,'base64'),
            captcha: verify,
            is_checked: is_checked.value,
            language: language
        };
        $.ajax.request({
            url: '/login_submit',
            method: "post",
            data: data,
            timeout: 30000,
            error: function () {
            },
            success: function (data) {
                if (data == "success") {
                    window.location.href = "/";
                    return;
                } else if (data == '1' || data == '2' || data == '3' || data == '6') {
                    window.location.href = "/modify";
                } else if (data == '8') {
                    showMsg(lang.countError, tip[0]);
                } else if (data == '99') {
                    showMsg(lang.verificationError, tip[1]);
                } else if (data == '4') {
                    showMsg(lang.addressLocked, tip[0]);
                } else if (data == '5') {
                    showMsg(lang.accountLocked, tip[0]);
                } else if (data == 'connect.error') {
                    showMsg(lang.maintenance, tip[0]);
                }
                change_captcha();
            }
        });
    }

    function is_form_submit() {
        if (is_checked.checked && is_checked.value == 'on') {
            $('#btn_submit').setAttribute('disabled', 'true');
            form_submit();
        } else {
            $('.tooltip')[0].style.display = 'block';
        }
    }

    //事件
    captcha.onkeydown = function (e) {
        $('#btn_submit').removeAttribute('disabled');
        if (e.keyCode == 13) {
            is_form_submit();
        }
    };

    is_checked.onclick = function () {
        if (this.checked) {
            this.value = 'on';
            $('.tooltip')[0].style.display = 'none';
            $('#btn_submit').className += ' btn_hover';
        } else {
            this.value = 'off';
            $('#btn_submit').className = 'login_btn';
        }
    };
    name.onkeydown = function (e) {
        $('#btn_submit').removeAttribute('disabled');
        if (e.keyCode == 32) {
            return false;
        }
        return true;
    };
    name.onkeyup = function (e) {
        $('#btn_submit').removeAttribute('disabled');
        var k = e.keyCode;
        if (k == 32 || k == 116) {
            /* space f5 */
            return false;
        } else if (k == 13) {
            /* enter */
            pwd.focus();
        } else if (k != 9) {
            /* tab */
            checkInfo();
        }
    };
    pwd.oninput = function (e) {
        checkInfo();
    };
    pwd.onkeydown = function (e) {
        $('#btn_submit').removeAttribute('disabled');
        if (e.keyCode == 32) {
            return false;
        }
        return true;
    };
    pwd.onkeyup = function (e) {
        $('#btn_submit').removeAttribute('disabled');
        var k = e.keyCode,
            pass = true;
        if (k == 13) {
            $("#btn_submit").focus();
        } else if (k == 32) {
            return false;
        }
    };
    $('#captcha_img_id').onclick = function () {
        change_captcha();
    };
    //辅助性功能函数
    function checkInfo() {
        if (name.value == "") {
            showMsg(lang.pleaseName, tip[0]);
            error = true;
        } else if (pwd.value == "") {
            showMsg(lang.pleasePW, tip[0]);
        } else {
            hideMsg();
        }
        return;
    }

    function showMsg(msg, dom) {
        dom.innerHTML = msg;
        dom.style.visibility = 'visible';
    }

    function hideMsg() {
        $(".tip")[0].style.visibility = "hidden";
    }

    function change_captcha() {
        var url = "/verify/random=" + Math.random();
        $("#captcha_img_id").setAttribute("src", url);
    }

    function clearShowMsg() {
        tip[0].innerHTML = '';
        tip[1].innerHTML = '';
    }

    var modal = {
        openModal: function () {
            var modal = $('.modal')[0],
                modal_body = $('.modal_body')[0];
            modal_body.innerHTML = "";
            var userLicenseHTML = "";
            if (typeof language == 'undefined') {
                userLicenseHTML = lang.declaration;
            } else {
                userLicenseHTML = language.declaration;
            }
            modal_body.insertAdjacentHTML('beforeend', userLicenseHTML);
            modal.style.display = 'block';
            mask.style.display = 'block';
        },
        closeModal: function () {
            var modal = $('.modal')[0];
            modal.style.display = 'none';
            mask.style.display = 'none';
        }
    };

    function agreeChecked() {
        is_checked.checked = 'true';
        is_checked.value = 'on';
        modal.closeModal();
        $('.tooltip')[0].style.display = 'none';
        $('#btn_submit').className += ' btn_hover';
    }


    //language
    function language_m_over() {
        $('.language_list_text')[0].style.display = 'block';
    }

    function language_m_out() {
        $('.language_list_text')[0].style.display = 'none';
    }

    function language_tab(obj) {
        $('.language_title')[0].innerText = obj.innerHTML;
        window.location.href = "/login.html?lang=" + obj.getAttribute('value');
    }

    function language_replace() {
        name.placeholder = lang.pleaseName;
        pwd.placeholder = lang.pleasePW;
        captcha.placeholder = lang.pleaseCAPTCHA;
        $('#btn_submit').value = lang.loginnow;
        $('#a_text_readUserLicense').innerHTML = lang.readAndAgreeLicense;
        $('#modal_title').innerHTML = lang.userLicense;
        $('#div_readLicense').innerHTML = lang.readLicense;
        $('#btn_text_agree').value = lang.btnAgreement;
        $('#copyright').innerHTML = lang.copyright;
    }

    language_replace();

    function crypt($str, $type) {
        switch ($type) {
            case 'base64':
                return codeHandler.encode($str, 'base64').replace(/\+/g, '%2B');
                //return enbase64(escape($str), 'base64').replace(/\+/g, '%2B');
            case 'rsa':
                var encrypt = new JSEncrypt();
                var pubkey = 'LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDckhCSnhDQWJ1dDhCL2phVmplWVZ3ckNjQgpRTERIak5XQVU3ZSt6WjluVWdDc0gza081bXkrYkhDR2pWUWkvTTcxMG5VYjBCWERkbFVKYm1PQndJNmU2NlQvCllxYnVmR3hFUUIxbjh0UXc1M0Rydkh4SGFyaFAyVGhhZ1pNQkRmcTk1MU92enpGWlZsUWYrSE9kV2pHd1EvQVYKUUFSelB1aWszc3NlWmFROFl3SURBUUFCCi0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQ==';
                encrypt.setPublicKey(codeHandler.decode(pubkey, 'base64'));
                return encrypt.encrypt($str).replace(/\+/g, '%2B');
        }
    }
    return {
        is_form_submit: is_form_submit,
        modal: modal,
        agreeChecked: agreeChecked,
        language_m_over: language_m_over,
        language_m_out: language_m_out,
        language_tab: language_tab,
        language_replace: language_replace
    };
}());

