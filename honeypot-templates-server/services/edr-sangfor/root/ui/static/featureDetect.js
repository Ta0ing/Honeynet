/**
 * @description 罍��ユ�頵��������弱�Vue鐚�篁ュ�ie羌頵������ゆ�㍼��Û�弱���ie11鐚�
 * @author lh
 * @date 2018.4.10
 */
 (function  () {
    var isIE = !!window.ActiveXObject || 'ActiveXObject' in window;
    var isIE11 = navigator.userAgent.indexOf('Trident') > -1 && navigator.userAgent.indexOf("rv:11.0") > -1;
    if (!detectSupportVue() || (isIE && !isIE11)) {    
        window.location.href="/ui/browser.php";
    }

    function detectSupportVue () {
        var obj = {},
            isSupport = false;
        try {
            Object.defineProperty(obj, 'a', {
                get: function () {
                    isSupport = true;
                }
            })
        } catch (e) {

        }
        obj.a;
        return isSupport;
    }
})();



