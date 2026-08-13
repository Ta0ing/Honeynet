// JavaScript Document
(function($){

$.su.Widget("password", {
	defaults: {
		fieldLabel: "",
		tips: "",
		hint: null, //$.su.CHAR.OPERATION.PASSWORD_HINT,
		validateIcon: false,
		showLevel: true,
		allowBlank: true,

		encrypt: $.su.encrypt,
		n: null,
		e: null,

		_minLength: 2,
		_maxLength: 16,

		vtype: "password",

		invalidText: $.su.CHAR.VTYPETEXT.INVALIDTEXT,
		blankText: $.su.CHAR.VTYPETEXT.BLANKTEXT,

		allowVisible: false
	},
	create: function(defaults, options){
		var me = this;
		me.each(function(i, obj){
			var tar = $(obj);
			$.extend(obj, defaults, options);

			

			var inHTML =	"<div class=\"container widget-container text-container password-container "+obj.cls+" "+(obj.showLevel ? "level" : "")+"\">";

			//fieldLabel
			if (obj.fieldLabel !== null){
				inHTML +=		"<div class=\"widget-fieldlabel-wrap "+obj.labelCls+"\">";
				inHTML +=			"<label class=\"widget-fieldlabel text-fieldlabel password-fieldlabel\">"+obj.fieldLabel+"</label>";

				if (obj.fieldLabel !== ""){
					inHTML +=		"<span class=\"widget-separator\">"+obj.separator+"</span>";
				};

				inHTML +=		"</div>";				
			};
				inHTML +=		"<div class=\"widget-wrap-outer text-wrap-outer password-wrap-outer\">"
				inHTML +=			"<div class=\"widget-wrap text-wrap password-wrap "+(obj.allowVisible ? "allow-visible" : "")+"\">";
				inHTML +=				"<span class=\"text-wrap password-wrap \">";
				inHTML +=					"<input type=\"password\" class=\"text-text password-text password-hidden "+obj.inputCls+"\" maxlength=\""+obj._maxLength+"\"/>";

			if (obj.allowVisible){
				obj.passwordVisible = false;

				inHTML +=					"<input type=\"text\" class=\"text-text password-text password-visible hidden "+obj.inputCls+"\" maxlength=\""+obj._maxLength+"\"/>";
				inHTML +=					"<span class=\"icon allow-visible-btn\"></span>";
			};

				inHTML +=				"</span>";

			if (obj.hint){
				inHTML +=				"<span class=\"hint text-hint password-hint\">";
				inHTML +=					"<input class=\"text-hint password-hint "+obj.inputCls+"\" value=\""+obj.hint+"\" contenteditable=\"false\" readonly=\"readonly\"/>";
				inHTML +=				"</span>";
			};

			if (obj.showLevel){
				inHTML +=			"<div class=\"password-level "+obj.inputCls+"\">";
				inHTML +=				"<span class=\"level low\">"+$.su.CHAR.OPERATION.LOW+"</span>";
				inHTML +=				"<span class=\"level middle\">"+$.su.CHAR.OPERATION.MIDDLE+"</span>";
				inHTML +=				"<span class=\"level high\">"+$.su.CHAR.OPERATION.HIGH+"</span>";
				inHTML +=			"</div>";
			};

				inHTML +=			"</div>";

			if (this.validateIcon){
				inHTML +=			"<span class=\"widget-validate-icon\"></span>";
			};

			if (this.tips){
				inHTML +=			"<div class=\"widget-tips textbox-tips "+obj.tipsCls+"\">";
				inHTML +=				"<div class=\"content tips-content\"></div>";
				inHTML +=			"</div>";
			};

				inHTML +=			"<div class=\"widget-error-tips textbox-error-tips "+obj.errorTipsCls+"\">";
				inHTML +=				"<span class=\"widget-error-tips-delta\"></span>";
				inHTML +=				"<div class=\"widget-error-tips-wrap\">";
				inHTML +=					"<div class=\"content error-tips-content\"></div>";
				inHTML +=				"</div>";
				inHTML +=			"</div>";

				inHTML +=		"</div>";
				inHTML +=	"</div>";

			var container = $(inHTML);
			tar.replaceWith(container);
			container.find("span.password-wrap").append(tar.addClass("hidden"));


			if (this.vtype){
				var vtype = this.vtype;

				if (this.vtypeText){
					this.vtype.vtypeText = this.vtypeText;
				};
				this.vtype = new $.su.vtype(vtype);
			};

		});


		//事件监听部分
		var container = me.closest("div.password-container"),
			//passwordText = container.find("input.password-text"),
			passwordHiddenText = container.find("input.password-hidden"),
			passwordVisibleText = container.find("input.password-visible"),
			passwordHint = container.find("input.password-hint"),
			passwordLevel = container.find("input.password-level");

		container.delegate("input.password-text", "click", function(e){
			e.stopPropagation();
		}).delegate("span.allow-visible-btn", "click", function(e){
			var obj = me.get(0);
			var btn = $(this);
			var text = "";
			if (obj.passwordVisible){
				text = passwordVisibleText.val();
				passwordHiddenText.val(text);

				passwordVisibleText.css("display", "none");
				passwordHiddenText.css("display", "inline-block");

				btn.removeClass("visible");
				obj.passwordVisible = false;
			}else{
				text = passwordHiddenText.val();
				passwordVisibleText.val(text);

				passwordHiddenText.css("display", "none");
				passwordVisibleText.css("display", "inline-block");

				btn.addClass("visible");
				obj.passwordVisible = true;
			};
		});

		me.on("ev_validatechange", function(e, isvalid, tips){
			if (isvalid){
				$(this).textbox("setValid");
			}else{
				$(this).textbox("setError", tips);
			};
		});

		//隐藏提示语句
		passwordHint.on("click focus", function(e){
			e.preventDefault();
			var obj = me.get(0);
			$(this).fadeOut(50, function(){
				if (obj.passwordVisible){
					passwordVisibleText.focus()
				}else{
					passwordHiddenText.focus();
				};
				//passwordText.select();
			});
			return false;
		});

		//passwordText.attr("contentEditable", true);
		container.delegate("input.password-text", "click focus", function(e){
			e.preventDefault();
			passwordHint.fadeOut(50);

			me.password("setFocus");
		});

		container.delegate("input.password-text", "blur", function(e){
			// var	obj = me.get(0),
			// 	encrypt = obj.encrypt,
			// 	e = obj.e,
			// 	n = obj.n;

			// var	value = obj.passwordVisible ? passwordVisibleText.val() : passwordHiddenText.val();

			// me.password("removeFocus");

			// if (!value){
			// 	passwordHint.css("display", "inline");
			// };
			// //console.log(me, encrypt, e, n)
			// if (encrypt && e && n){
			// 	value = encrypt(n, e, value);
			// };

			// me.val(value);
			me.password("doEncrypt");
		});

		container.delegate("input.password-text", "keyup", function(e){
			//这里作规则检查
			var value = $(this).val();
			me.password("levelCheck");
			me.password("validate");
			me.trigger("ev_change", [value, e.key, e.keyCode]);
		});

		//初始化
		setTimeout(function(){	//修正因为浏览器记住密码产生的bug
			if (passwordHiddenText.val() != ""){
				passwordHint.css("display", "none");
			};
		}, 100);

		me.password("setTips", options.tips);
		
		return me;
	},/*,
	disable: function(me){
		var me = me || this;

	},
	enable: function(me){
		var me = me || this;

	},*/
	levelCheck: function(me){
		var me = me || this,
			obj  = me.get(0),
			container = me.closest("div.password-container"),
			passwordVisible = container.find("input.password-visible"),
			passwordHidden = container.find("input.password-hidden"),
			value = "";

		if (obj.passwordVisible){
			value = passwordVisible.val();
		}else{
			value = passwordHidden.val();
		};

		/*var lv = 0;
	    if(value.match(/[a-z]/g)){lv++;};
	    if(value.match(/[0-9]/g)){lv++;};
	    if(value.match(/(.[^a-z0-9])/g)){lv++;};
	    if(value.length < 6){lv=0;};
	    if(lv > 3){lv=3;};

	    // 0 危险， 1、low 2、middle 3、high
	    var passwordLevel = container.find("div.password-level");
	    passwordLevel.removeClass("level-0 level-1 level-2 level-3");
	    passwordLevel.fadeIn(100).addClass("level-"+lv);*/

	    var point = 0;
	    var charLength = value.length;
	    if (0 < charLength && charLength <= 4){
	    	point += 5;
	    }else if (4 < charLength && charLength <= 7){
	    	point += 10;
	    }else if (charLength > 7){
	    	point += 25;
	    };

	    var patternLowCase = /[a-z]/g;
	    var patternUpCase = /[A-Z]/g;
	    var patternNum = /[0-9]/g;
	    var patternSign = /[\`\~\!\@\#\$\%\^\&\*\(\)\-\=\_\+\[\]\{\}\;\:\'\"\\\|\/\?\.\,\<\>\x20]/g;

	    var flagLow = patternLowCase.test(value);
	    var flagUp = patternUpCase.test(value);

	    var flagChar = flagLow || flagUp;
	    var flagBoth = flagLow && flagUp; 

	    if (flagBoth){
	    	point += 20;
	    }else if (flagChar){
	    	point += 10;
	    };

	    var flagNum = patternNum.test(value);
	    var len = value.match(patternNum);
	    if (len){
	    	len = len.length;
	    }else{
	    	len = 0;
	    };
	    if (len >= 3){
	    	point += 20;
	    }else if (0 < len && len < 3){
	    	point += 10;
	    };

	    var flagSign = patternSign.test(value);
	    var len = value.match(patternSign);
	    if (len){
	    	len = len.length;
	    }else{
	    	len = 0;
	    };
	    if (len > 1){
	    	point += 20;
	    }else if (len == 1){
	    	point += 10;
	    };

	    if (flagBoth && flagNum && flagSign){
	    	point += 5;
	    }else if (flagChar && flagNum && flagSign){
	    	point += 3;
	    }else if (flagChar && flagNum){
	    	point += 2;
	    };

	    //console.log(value, charLength, "小写字母：",flagLow, "大写字母：",flagUp, "数字：", flagNum, "符号：",flagSign, point);

	    var lv = 0;
	    if (point <= 0){
	    	//灰
	    	lv = 0;
	    }else if (0 < point && point <= 50){
	    	//low
	    	lv = 1;
	    }else if (50 < point && point < 80){
	    	//middle
	    	lv = 2;
	    }else{
	    	//high
	    	lv = 3;
	    };

	    var passwordLevel = container.find("div.password-level");
	    passwordLevel.removeClass("level-0 level-1 level-2 level-3");
	    passwordLevel.fadeIn(100).addClass("level-"+lv);

	    return me;
	    
	},
	validate: function(me, flag){
		var me = me || this,
			obj = me.get(0),
			container = me.closest("div.password-container"),
/*			password = container.find("input.password-text"),
			value = password.val(),	//不trim！！！！！*/
			passwordVisible = container.find("input.password-visible"),
			passwordHidden = container.find("input.password-hidden"),
			value = "",
			result = true,
			flag = (flag[1] === false) ? false : true;

		if (obj.passwordVisible){
			value = passwordVisible.val();
		}else{
			value = passwordHidden.val();
		};

		var returnResult = function(tar, resultText){
			if (resultText === true){
				if (flag){
					$(tar).trigger("ev_validatechange", [true, tar.tips]);
				};
				return true;
			}else{
				if (!resultText){
					resultText = tar.tips;
					//console.log("textbox no error tips!", tar.name);
				};
				if (flag){
					$(tar).trigger("ev_validatechange", [false, resultText]);
				};
				return false;
			};
		};

		//console.log("1", result);
		/*if (obj.allowBlank !== true){
			if (value == ""){
				result = false;
				return returnResult(obj, result, obj.blankText);
			}else{
				result = true;
			};
		};*/

		//console.log("2", result);
		//console.log(value, obj.lengthPattern, $.su.unicode.encodeNoPrefix(value), obj.lengthPattern.test($.su.unicode.encodeNoPrefix(value)))

		if (value == ""){
			if (obj.allowBlank !== true){
				result = false;
				return returnResult(obj, obj.blankText);
			}else{
				result = true;
				return returnResult(obj, true);
			};
		}else{
			/*if (!$.su.unicode.lengthCheck(value, obj._minLength, obj._maxLength)){
				result = false;
				return returnResult(obj, result, obj.invalidText);
			};*/

			var vtype = obj.vtype;
			//console.log(vtype)
			if (vtype && vtype.isVtype === true){
				result = vtype.validate(value);
				//console.log("run", result)
				if (result !== true){
					if (result === false){
						return returnResult(obj, vtype.vtypeText);
					}else{
						return returnResult(obj, result);
					};
				};
			};

			if (obj.validator){
				result = obj.validator(value);
				if ($.type(result) !== "boolean" || result == false){
					result = false;
					return returnResult(obj, obj.invalidText);
				};
			};
			//console.log("3", result)
			return returnResult(obj, true);
		};
	},
	setValue: function(me, _value){
		var me = me || this,
			obj = me.get(0),
			value = _value[1];
		//console.log($.type(me.get(0).encrypt))
		if ($.type(me.get(0).encrypt) == "function" && $.isArray(value)){
			obj.n = value[0];
			obj.e = value[1];
		}else{
			me.val(value);
		};

		/*if (obj.hint){
			me.closest("div.text-container").find("span.hint").css("display", "none");
		};*/

		//me.attr("snapshot", value);
		//me.val(value);

		return me;
	},
	getValue: function(me){
		var me = me || this,
			container = me.closest("div.password-container"),
			password = container.find("input.password-text");

		return password.val();
	},
	doEncrypt: function(me){
		var	obj = me.get(0),
			encrypt = obj.encrypt,
			e = obj.e,
			n = obj.n;

		var container = me.closest("div.password-container"),
			//password = container.find("input.password-text"),
			passwordHiddenText = container.find("input.password-hidden"),
			passwordVisibleText = container.find("input.password-visible"),
			passwordHint = container.find("input.password-hint");

		var	value = obj.passwordVisible ? passwordVisibleText.val() : passwordHiddenText.val();

		me.password("removeFocus");

		if (!value){
			passwordHint.css("display", "inline");
		};
		//console.log(me, encrypt, e, n)
		if (encrypt && e && n){
			value = encrypt(n, e, value);
		};

		me.val(value);

		return me;
	},
	disable: function(me){
		var me = me || this,
			container = me.textbox("getContainer"),
			password = container.find("input.password-text");

		container.addClass("disabled");
		password.prop("disabled", true);

		return me;
	},
	enable: function(me){
		var me = me || this,
			container = me.textbox("getContainer"),
			password = container.find("input.password-text");

		container.removeClass("disabled");
		password.prop("disabled", false);

		return me;
	}
});

})(jQuery);