/**
 * 提示板面组件
 * 调用方法 jQuery(tar).TipPanel({ ... });
 * @param options 配置项 @link $.fn.TipPanel.defaults
 * @version 1.0
 */
(function($) {

    var __triFlagClass = '__tippanel__gen__',
        rubbishs = []; //回收对象

    $.fn.TipPanel = function(options) {

        //将用户配置的配置项覆盖默认配置值
        var opts = $.extend({}, $.fn.TipPanel.defaults, options),
            //触发打开面板的按钮配置项
            trigOpts = opts.trigger,
            id = 1;

        function mate(evt) {
            return $(evt.target).hasClass(__triFlagClass) ||
                $(evt.target).parents().hasClass(__triFlagClass) && opts.filter(evt);
        }

        function toShow(evt) {
            //对某些节点执行click时需要判断
            if (!mate(evt)) {
                return true;
            }
            var titTar, body, lastPanel,
                scope = opts.scope,
                tpl = opts.tpl,
                arrow, first = true,
                panel = $(this).data('act.TP'),
                //关闭板面的方式
                isHide = opts.closeAction === 'hide',
                autoShow = opts.autoShow !== false;
            this.autoHideTimer;
            //已经打开了，就直接关闭
            if ($(this).data('show')) {
                //不自动关闭的类型，多次触发都不会变
                if (noAutoHide.indexOf(opts.triggerType) !== -1) {
                    return;
                }
                setTrigStateCls(this, trigOpts, false);
                close(panel, isHide, this, trigOpts)();
                return;
            }

            setTrigStateCls(this, trigOpts, true);
            //标识已经显示
            autoShow && $(this).data('show', true);

            //如果还没创建
            if (!panel) {

                panel = $($.fn.TipPanel.template.replace(/\{cls\}/, opts.cls));
                //绝对定位，可以随意移位
                panel.css({
                    position: 'absolute'
                });
                panel.opts = opts;
                panel.itrigger = this;
                panel.attr('id', 'tippanel_gen_' + id++);
                panel.tpclose = (function(panel, isHide, trigger, trigOpts) {
                    return function(cb) {
                        close(panel, isHide, trigger, trigOpts)(cb);
                    }
                })(panel, isHide, this, trigOpts);
                panel.tpshow = toShow;
                panel.toshow = function() {
                    panel.css({
                        display: 'block',
                        visibility: 'visible'
                    });
                }
                //将缓存到该目标对象下，下次则不能重新渲染了
                $(this).data('act.TP', panel);
            } else {
                first = false;
            }
            //清理掉最后打开的一个界面
            lastPanel = $.fn.TipPanel.lastPanel;
            //同一个面板就不用去执行
            if (lastPanel && lastPanel !== panel) {
                close(
                    lastPanel,
                    lastPanel.opts.closeAction === 'hide',
                    lastPanel.itrigger,
                    lastPanel.opts.trigger
                )();
            }
            $.fn.TipPanel.lastPanel = panel;

            //有配置title或配置close,暂时禁用
            if (false && (opts.title || opts.closable === true)) {
                titTar = panel.find('.tippanel-title');
                //标题块可见
                titTar.css({
                    display: 'block'
                });
                //可显示标题
                if (opts.title) {
                    titTar.find('.tp-title').html(opts.title);
                }
                //可手动关闭
                if (opts.closable === true) {
                    titTar.find('.tp-close').css({
                        display: 'block'
                    }).click(close(panel, isHide, this, trigOpts));
                }
            }
            //渲染在页面上
            //opacity = 0;for ie
            panel.css({
                display: 'block',
                visibility: 'hidden',
                zIndex: zIndex++,
                opacity: 0
            });
            //第一次显示时，才需要append到document，并更新内容,初始化脚本
            //如果是close的，也需要重新插到document上
            if (first || !isHide) {
                panel.appendTo(opts.renderTo && $(opts.renderTo) || document.body);
            }

            body = panel.find('.tippanel-content');
            if (first || opts.reset) {
                //暂定为最简单的模板
                //到时再改成可以动态数据生成的模板
                if (tpl) {
                    if ($.isFunction(tpl)) {
                        tpl = tpl.call(scope || window, panel, this);
                    }
                    body.html(tpl);
                }
            }
            if (first || opts.reset || !isHide) {
                //执行初始化逻辑
                opts.already.call(scope || window, panel, this);
            }
            if (opts.css) {
                body.css(opts.css);

            }
            var offsets = refix(panel, this);
            panel.css({
                top: offsets.top,
                left: offsets.left,
                display: 'block',
                visibility: autoShow ? 'visible' : 'hidden '
            });
            //显示效果
            if (opts.fade) {
                panel.animate({
                    opacity: 1
                });
            } else {
                panel.css({
                    opacity: 1
                });
            }
            //延时自动关闭
            if (opts.autoClose) {
                this.autoHideTimer = setTimeout(
                    close(panel, isHide, this, trigOpts),
                    opts.dismissDelay
                );
            }
        }

        function toClose(evt) {
            var panel = $(this).data('act.TP');
            if (!mate(evt) || !panel) {
                return true;
            }
            close(
                panel,
                panel.opts.closeAction === 'hide',
                panel.itrigger,
                panel.opts.trigger
            )();
        }
        return this.each(function() {
            var trig = this,
                oncls = trigOpts.hoverCls;
            //去掉鼠标点击的外框
            $(this).css({
                outline: 'none'
            }).addClass(__triFlagClass);
            //鼠标悬浮事件
            if (oncls) {
                $(this).hover(
                    function(e) {
                        $(this).addClass(oncls);
                    },
                    function(e) {
                        //如果板面处于打开状态，则不执行
                        if (!$(trig).data('show')) {
                            $(this).removeClass(oncls);
                        }
                    }
                );
            }
            //鼠标事件
            switch (opts.triggerType) {
                case 'hover':
                    $(this).hover(function(evt) {
                        var me = this;
                        this.st = setTimeout(function() {
                            toShow.call(me, evt);
                        }, 500);
                    }, function(evt) {
                        clearTimeout(this.st);
                        toClose.call(this, evt);
                    });
                    break;
                case 'form':
                    toShow.call(this, {
                        target: this
                    });
                    opts.autoShow = true;
                    $(this).mouseover(toShow);
                    break;
                case 'show':
                    toShow.call(this, {
                        target: this
                    });
                    opts.triggerType = 'click';
                default:
                    $(this)[opts.triggerType](toShow);
            }
            rubbishs.push(trig);
        });
    };
    var zIndex = 99999,
        dir = ['t-r', 't-l', 'b-r', 'b-l', 'l', 'r'],
        noAutoHide = ['form'],
        wh, ww;
    $.fn.TipPanel.lastPanel; //一个页面只能一个tip
    /**
     * 关闭板面，有两种方式，close[hide == false]或hide[hide == true]
     * 默认为close，会清除DOM节点
     */
    function close(panel, hide, trigger, trigOpts) {
        return function(cb) {
            panel.fadeOut(panel.opts.fade ? 200 : 0, function() {
                if (!hide) {
                    panel.remove();
                    // $(trigger).removeData('act.TP');
                    //	delete panel.itrigger;
                    //	delete panel.tpclose;
                    //	delete panel.tpshow;
                    //	delete panel.toshow;
                }
                $(trigger).data('show', false);
                //恢复触发按钮状态
                setTrigStateCls(trigger, trigOpts, false);
                //清理定时器
                if (trigger.autoHideTimer) {
                    clearTimeout(trigger.autoHideTimer);
                }
                if (cb) {
                    cb();
                }
            });
        }
    }
    /**
     * 设置触发按钮的样式，可配置为activeCls,其次为hoverOn
     */
    function setTrigStateCls(trigger, trigOpts, add) {
        $(trigger)[add ? 'addClass' : 'removeClass'](trigOpts.activeCls || trigOpts.hoverCls);
    }
    //自动定义只有上下方位之左右分。
    function getGravityDire(gravity, pos, width, height, padding) {
        //这两个值不是固定的，所以在调用时再获取
        wh = $(document.documentElement).height();
        ww = $(document.documentElement).width();
        var wrap = [pos.top, //top-wrap
                ww - pos.left - pos.width, //right-wrap 
                wh - pos.top - pos.height, //bottom-wrap
                pos.left
            ], //left-wrap
            topWrap = wrap[0] - wrap[2] > 0 || wrap[0] >= height,
            rightWrap = wrap[1] - wrap[3] > 0 || wrap[1] >= width,
            result, match;
        //如果指定了位置，且配置了自动定位，
        //则要检查用户指定的位置是否够显示，不然就进行全自动定位
        if (gravity) {
            switch (gravity) {
                case 't-l':
                    match = wrap[0] >= height && (wrap[3] + pos.width / 2 > width);
                    break;
                case 't-r':
                    match = wrap[0] >= height && (wrap[1] + pos.width / 2 > width);
                    break;
                case 'b-l':
                    match = wrap[2] >= height && (wrap[3] + pos.width / 2 > width);
                    break;
                case 'b-r':
                    match = wrap[2] >= height && (wrap[1] + pos.width / 2 > width);
                    break;
                case 'l':
                    match = wrap[3] >= width;
                    break;
                case 'r':
                    match = wrap[1] > width;
                    break;
            }
            if (match) {
                return gravity;
            }
        }

        if (topWrap && rightWrap) {
            result = dir[0];
        } else if (topWrap && !rightWrap) {
            result = dir[1];
        } else if (!topWrap && rightWrap) {
            result = dir[2];
        } else if (!topWrap && !rightWrap) {
            result = dir[3];
        }
        return result;
    }
    /**
     * 调整面板的最终位置
     */
    function fixPosition(gravity, pos, actWidth, actHeight, posfix, autoGravity, padding, arrow) {

        var top, //根据DOM结构计算出来的top值
            left, //根据DOM结构计算出来的left值 
            actTop, //最终根据位移计算出来的top值
            actLeft, //最终根据位移计算出来的left值
            tbPadding = padding.top + padding.bottom,
            lrPadding = padding.left + padding.right;
        //自动定位
        if (autoGravity) {
            gravity = getGravityDire(gravity, pos, actWidth + lrPadding, actHeight + tbPadding, padding);
        }
        switch (gravity) {
            case 't-l':
                top = pos.top - actHeight + 4;
                left = pos.left + pos.width / 2 - actWidth + padding.right + 17;
                arrow.className = 'arrow arrow_bottom arrow_tb_right';
                break;
            case 't-r':
                top = pos.top - actHeight + 4;
                left = pos.left + pos.width / 2 - padding.left - 19;
                arrow.className = 'arrow arrow_bottom arrow_tb_left';
                break;
            case 'b-l':
                top = pos.top + pos.height;
                left = pos.left + pos.width / 2 - actWidth + padding.right + 17;
                arrow.className = 'arrow arrow_top arrow_tb_right';
                break;
            case 'b-r':
                top = pos.top + pos.height;
                left = pos.left + pos.width / 2 - padding.left - 19;
                arrow.className = 'arrow arrow_top arrow_tb_left';
                break;
            case 'l':
                top = pos.top + pos.height / 2 - padding.top - 19;
                left = pos.left - actWidth;
                arrow.className = 'arrow arrow_right';
                break;
            case 'r':
                top = pos.top + pos.height / 2 - padding.top - 19;
                left = pos.left + pos.width;
                arrow.className = 'arrow arrow_left';
                break;
        }
        actLeft = left + posfix.left;
        actTop = top + posfix.top;

        return {
            top: actTop,
            left: actLeft,
            anchorDir: gravity
        };
    }

    function refix(panel, trigger) {
        var pos = $.extend({}, $(trigger).offset(), {
                width: trigger.offsetWidth,
                height: trigger.offsetHeight
            }),
            //面板的实际大小
            width = panel[0].offsetWidth,
            height = panel[0].offsetHeight,
            opts = panel.opts,
            //面板的内空白
            padding = {
                top: parseInt(panel.css('paddingTop')),
                left: parseInt(panel.css('paddingLeft')),
                bottom: parseInt(panel.css('paddingBottom')),
                right: parseInt(panel.css('paddingRight'))
            },
            arrow = $('div.arrow', panel),
            //位移偏值
            posfix = $.extend({
                left: 0,
                top: 0
            }, opts.posfix);
        return fixPosition(
            opts.gravity,
            pos,
            width,
            height,
            posfix,
            opts.autoGravity,
            padding,
            arrow[0]
        );
    }
    $(window).resize(function() {
        var lastPanel = $.fn.TipPanel.lastPanel,
            pos;
        if (!lastPanel || !$(lastPanel.itrigger).data('show')) {
            return;
        }
        pos = refix(lastPanel, lastPanel.itrigger);
        lastPanel.css({
            top: pos.top,
            left: pos.left
        });
    }).unload(function() {
        $.each(rubbishs, function(index, obj) {
            $(obj).removeData().unbind();
        })
    });

    //点击其他地方可关闭板面，实现有点-_-
    $(document).click(function(evt) {
        var lastPanel = $.fn.TipPanel.lastPanel,
            tar;
        //如果不存在或非可视状态下直接返回
        if (!lastPanel || !$(lastPanel.itrigger).data('show')) {
            return;
        }
        tar = $(evt.target);
        if (tar[0].id != lastPanel[0].id &&
            tar.parents('#' + lastPanel[0].id).length == 0 &&
            !tar.hasClass(__triFlagClass) &&
            !tar.parents().hasClass(__triFlagClass) &&
            noAutoHide.indexOf(lastPanel.opts.triggerType) === -1
        ) {
            close(
                lastPanel,
                lastPanel.opts.closeAction === 'hide',
                lastPanel.itrigger,
                lastPanel.opts.trigger
            )();
        }
    });
    //模板
    /*$.fn.TipPanel.template = [
    	'<div class="tippanel {cls}">',
    		'<div class="tippanel-inner">',
    			'<div class="tippanel-title">',
    				'<span class="tp-title"></span>',
    				'<span class="tp-close"></span>',
    			'</div>',
    			'<div class="tippanel-content"></div>',
    		'</div>',
    	'</div>'].join('');*/
    $.fn.TipPanel.template = [
        '<div class="tippanel {cls}">',
        '<div class="arrow"></div>',
        '<div class="tippanel-inner">',
        '<div class="tippanel-content"></div>',
        '</div>',
        '</div>'
    ].join('');
    //默认配置
    $.fn.TipPanel.defaults = {
        filter: function() {
            return true;
        },
        //[option]tpl,already方法的scope，如果没有默认就是window
        scope: window,
        //提示内容，支持HTML,如果有脚本需要读取这个配置的dom节点，
        //务必把代码写在already接口中
        //如果是一个function,则参数为[tippanel, trigger]
        tpl: '',
        //[option]触发事件类型,不同事件类型会有意想不到的效果
        //show是另外一种形式，表示会自动显示，之后以点击再次显示
        //form　用于表单验证
        triggerType: 'click',
        //[option]标题
        title: null,
        //[option]自定义样式类名,如果有多个就用空格分开写就行了
        cls: '',
        //[option]针对tippanel-content的样式配置
        //如设置板面大小{height : 100, width : 100}
        css: {
            width: 300,
            height: 'auto'
        },
        //[option]是否关闭，默认false
        closable: false,
        //[option]是否启用显示/关闭的缓冲效果
        fade: true,
        //[option]关闭方式，直接清除DOM节点为close，直接隐藏则为hide
        closeAction: 'close',
        //[option]板面显示的位置，默认为t-r,[t-r,t-l, b-r, b-l, l, r]
        gravity: 't-r',
        //[option]是否在左右或上下自动定位。
        autoGravity: true,
        //[option]
        autoShow: true,
        //[option]是否自动关闭,默认false
        autoClose: false,
        //[option]自动关闭延时时间，默认5秒
        dismissDelay: 5000,
        //[option]每次显示都重置内容，对于非固定，需要改成true
        reset: false,
        //[option]板面的移动，默认方位都是居中
        posfix: {
            left: 0, //[option]+-
            top: 0 //[option]+-
        },
        //[option]触发按钮配置项
        trigger: {
            //[option]激活状态的样式类名
            activeCls: null,
            //[option]鼠标悬浮时的样式类名
            hoverCls: null
        },
        //[option]板面渲染后，初始化脚本的接口
        //参数为[tippanel, trigger]
        already: $.noop
    };
})(jQuery);