// 企业微信扫码登录获取二维码
function getCode() {
    document.querySelector('#loginMain').style.display = 'none';
    document.querySelector('.wxCode').style.display = 'flex';
    document.querySelector('.wxCode').style.opacity = 1;
    const redirect_uri = encodeURI('http://jhwx.bestlink.com.cn/cloudApp/weChatService/login/wx/index/48/0');
    window.WwLogin({
        "id": "wx_reg",
        "appid": "wx077cf5f19ddfdcc5",
        "agentid": 48,
        "redirect_uri": redirect_uri,
        "state": new Date().getTime(),
        "href": "",
    });
}

// 取消二维码登录
function closeCode() {
    document.querySelector('#loginMain').style.display = '';
    document.querySelector('.wxCode').style.display = 'none';
    document.querySelector('.wxCode').style.opacity = 0;
}

function getErrorMessage(keywordid, labelParams) {
    const language = 'zh_CN';
    if (!labelParams) {
        labelParams = [];
    }
    const labelname = jQuery.ajax({
        url: "/ServiceAction/com.eweaver.base.security.servlet.LoginAction?action=getLabelNameByKeyId",
        data: {keywordid: keywordid, language: language, labelParams: labelParams.toString()},
        async: false
    }).responseText;
    return labelname;
}


function login() {
    if (document.getElementById("uname").value && document.getElementById("j_password").value) {
        document.getElementById("j_username").value=document.getElementById("uname").value;
        return true;
    } else {
        if (document.getElementById("uname").value) {
            alert('请输入用户名');
        } else if (document.getElementById("j_password").value) {
            alert('请输入密码');
        }
    }
    return false;
}
