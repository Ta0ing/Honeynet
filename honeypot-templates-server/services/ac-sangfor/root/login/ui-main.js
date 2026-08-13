/*
 * Multilightbox for jQuery; all picture and effect defined on webconfig.xml;
 * author: benxshare (AT) gmail.com;
 */
(function($) {

	var __fmt = function(s, arr) {
		return s.replace(/%\d+/g, function(m) {
			var i = m.replace('%', '') * 1;
			return arr[i] === null ? '' : arr[i];
		});

	};

	$.fn.sfMultiLightBox = function(opt) {
		var me = this;
		this.hdl = $(this);
		this.startSlider = function(ani, delay) {
			this.hdl.attr('hdl', function() {
				return window.setInterval(function() {
					me.animateFunc(ani);
				}, delay);
			});
		};

		this.stopSlider = function() {
			var hdl = this.hdl.attr('hdl');
			window.clearInterval(hdl);
		};

		var fx = {
			slide: function($out, $in, dir) {
				var o = {};
				o[dir] = '0%';
				$($out).removeClass('ontop').addClass('layer-1').css(o);
				o[dir] = '100%';
				$($in).addClass('ontop').css(o);
				o[dir] = '-100%';
				$($out).animate(o, 1000);
				o[dir] = '0%';
				$($in).animate(o, 500, function() {
					$($out).removeClass('layer-1');
				});
			},
			$left: function($out, $in) {
				fx.slide($out, $in, 'left');
			},
			$right: function($out, $in) {
				fx.slide($out, $in, 'right');
			},
			$top: function($out, $in) {
				fx.slide($out, $in, 'top');
			},
			$bottom: function($out, $in) {
				fx.slide($out, $in, 'bottom');
			},
			boom: function($out, $in, x, y) {
				var o = {
					width: '0%', height: '0%'
				};
				$($out).removeClass('ontop').addClass('layer-1');
				o[x] = 0;
				o[y] = 0;
				$($in).addClass('ontop').css(o);
				$($in).animate({
					width: '100%', height: '100%'
				}, 500, function() {
					$($out).removeClass('layer-1');
				});
			},
			boom_right_top: function($out, $in) {
				fx.boom($out, $in, 'right', 'top');
			},
			boom_right_bottom: function($out, $in) {
				fx.boom($out, $in, 'right', 'bottom');
			},
			boom_left_top: function($out, $in) {
				fx.boom($out, $in, 'left', 'top');
			},
			boom_left_bottom: function($out, $in) {
				fx.boom($out, $in, 'left', 'bottom');
			},
			boom_center: function($out, $in) {
				$($out).removeClass('ontop').addClass('layer-1');
				$($in).addClass('ontop').css({
					top: '50%', left: '50%', width: '0%', height: '0%'
				});
				$($in).animate({
					width: '100%', height: '100%', top: '0%', left: '0%'
				}, 500, function() {
					$($out).removeClass('layer-1');
				});
			},
			slice_x: function($out, $in) {
				$($out).removeClass('ontop').addClass('layer-1');
				$($in).addClass('ontop').css({
					left: '50%', width: '0%'
				});
				$($in).animate({
					width: '100%', left: '0%'
				}, 500, function() {
					$($out).removeClass('layer-1');
				});
			},
			slice_y: function($out, $in) {
				$($out).removeClass('ontop').addClass('layer-1');
				$($in).addClass('ontop').css({
					top: '50%', height: '0%'
				});
				$($in).animate({
					height: '100%', top: '0%'
				}, 500, function() {
					$($out).removeClass('layer-1');
				});
			}
		};

		this.animateFunc = function(cfg) {
			//console.info(o);
			var arrs = cfg.items, l = arrs.length, idx = 0;
			var func = function(o) {
				var effect = o.effect || 'slideleft',
					$el = $(o.rel);
				var items = $el.find('a'), l = o.length;

				var current = parseInt($el.attr('idx'), 10);
				var next = (current >= l - 1) ? 0 : current + 1;
				$el.attr('idx', next);
				fx[effect](items[current], items[next]);
			};

			var delayFunc = function(x) {
				if(x.length > 1) func(x); //if this area images length greater than one, then effect it;
				idx++;
				if (idx < l) {
					window.setTimeout(function() {
						delayFunc(arrs[idx]);
					}, x.timeout);
				}
			};

			if (arrs[idx].timeout > 0) {

				window.setTimeout(function() {
					delayFunc(arrs[idx]);
				}, arrs[idx].timeout);

			} else {
				$.each(arrs, function(index, o) {
					func(o);
				});
			}
		};



		function createLightBox(xml) {
			var $xml = $(xml);
			var o = {items: [], timeout: null}, delay = 5000;
			var txt = $($xml.find('loading')[0]).context.textContent;
			$xml.find('lightbox area').each(function(index, area) {
				var efcs = ['$left', '$top', '$right', '$bottom', 'boom_left_top', 'boom_left_bottom', 'boom_right_top', 'boom_right_bottom', 'boom_center', 'slice_x', 'slice_y'];
				var el = $(area),
					rel = el.attr('rel'),
					efct = el.attr('effect') || efcs[parseInt(Math.random() * 11, 10)],
					timeout = el.attr('timeout') ? parseInt(el.attr('timeout'), 10) : el.attr('timeout'),
					arr = [], oa = {rel: rel, effect: efct, timeout: timeout};
				el.find('item').each(function(i, t) {
					var tpl = '<a %0 class="%1" target="%4"><img src="%2" /><span>%3</span></a>';
					var it = $(t);
					var url_str = '';
					if (it.attr('url')) {
						//配置中的url做了编码处理,则需要解码后使用
						var b = new Base64();
						var strs = it.attr('url').split("?");
						url_str = 'href="' + b.decode(strs[0]) + '?' + strs[1] + '"';
					}
					var src = it.attr('src'),
						text = it.attr('text'),
						url = url_str,
						css = (i === 0) ? 'ontop' : '',
						target = it.attr('target') || '_blank',
						tpx = __fmt(tpl, [url, css, src, text, target]);
					arr.push(tpx);
				});
				oa['length'] = arr.length;
				$(rel).html(arr.join('')).attr('idx', 0);
				$(rel).find('a').each(function(index, lnk) {
					$(this).find('span').fadeOut('fast');
					$(lnk).hover(function(obj) {
						$(this).find('span').fadeTo('fast', .6);
					}, function() {
						$(this).find('span').fadeTo('fast', 0);
					});
				});
				o['items'].push(oa);

				var papa = $(rel);
				papa.append('<div class="ldc">' + txt + '</div>');
				papa.find('.ldc').hide();
				papa.attr('ptr', window.setTimeout(function() {
					papa.find('.ldc').show();
				}, 2000));
			});
			me.startSlider(o, delay);

		}

		function sliceLightBoxText(xml) {

			var $xml = $(xml);
			var arrs = ['call', 'boxtitle', 'copyright'];
			$.each(arrs, function(idx, key) {
				$xml.find(key).each(function(i, t) {
					var el = $(t).eq(0), rel = el.attr('rel'),
						text = el.eq(0).context.textContent;
					$(rel).html(text);
				});
			});
		}

		$.get((opt.url || 'config.xml'), function(xml) {
			createLightBox(xml);

			sliceLightBoxText(xml);
		});



	};

})(jQuery);

function getcookie(name) {
	var cookie_start = document.cookie.indexOf(name);
	var cookie_end = document.cookie.indexOf(";", cookie_start);
	return cookie_start == -1 ? '' : unescape(document.cookie.substring(cookie_start + name.length + 1, (cookie_end > cookie_start ? cookie_end : document.cookie.length)));
}
function setCookie(name,value,time){
	var strsec = getsec(time);
	var exp = new Date();
	exp.setTime(exp.getTime() + strsec*1);
	document.cookie = name + '='+ escape (value) + ';secure=true;expires=' + exp.toGMTString();
}
/*
* 这是有设定过期时间的使用示例：
* s20是代表20秒
* h是指小时，如12小时则是：h12
* d是天数，30天则：d30
* 暂时只写了这三种
*/
function getsec(str){
	var num=str.substring(1,str.length)*1; 
	var str2=str.substring(0,1);
	var ret = 0;
	switch(str2.toLowerCase()){
		case 's':
		ret = num * 1000;
		break;
		case 'h':
		ret = num * 60 * 60 * 1000;
		break;
		case 'd':
		ret = num *24*60*60*1000;
		break;
	}
	return ret;
}


(function($) {
	$(document).ready(function() {
		/* 初始化广告 */
		$('div.gg-photo').sfMultiLightBox({
			url: 'login/config.xml'
		});

	});

	/* 检测大写锁定 */
	function detectCapsLock(ae) {
		var uO = ae.keyCode || ae.charCode,
			Uc = ae.shiftKey;
		if ((uO >= 65 && uO <= 90 && !Uc) || (uO >= 97 && uO <= 122 && Uc))
		{
			$(".shiftKey").slideDown();
		}
		else if ((uO >= 97 && uO <= 122 && !Uc) || (uO >= 65 && uO <= 90 && Uc))
		{
			$(".shiftKey").slideUp();
		}
		else
		{
			$(".shiftKey").slideUp();
		}
	}
})(jQuery);
