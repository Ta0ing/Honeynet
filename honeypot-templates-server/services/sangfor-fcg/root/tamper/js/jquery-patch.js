//给数组添加indexOf方法
if (!Array.prototype.indexOf) {
    Array.prototype.indexOf = function(o, from) {
        var len = this.length;
        from = from || 0;
        from += (from < 0) ? len : 0;
        for (; from < len; ++from) {
            if (this[from] === o) {
                return from;
            }
        }
        return -1;
    }
}