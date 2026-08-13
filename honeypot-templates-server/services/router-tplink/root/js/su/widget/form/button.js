// JavaScript Document
(function($){

$.su.Widget("button", {
	defaults: {
		text: "",
		//icon: "",
		iconCls: "",
		//type: "button", //ie不兼容定义type，所以需要直接写在html中！

		enableToggle: false,
		toggleHander: null,	//function(state)

		pressedCls: "pressed",
		pressed: false, /*/false 这个属性动态设置！*/
		
		cls: "",
		btnCls: "",

		/*loading: false,
		loadingText: $.su.CHAR.CONNECT.CONNECTING,
		loadedText: $.su.CHAR.CONNECT.CONNECTED,
		unloadingText: $.su.CHAR.CONNECT.DISCONNECTING,
		unloadedText: $.su.CHAR.CONNECT.DISCONNECTED,
		handler: null,*/

		fieldLabel: null
	},
	create: function(defaults, options){
		var me = this;
		me.each(function(i, obj){
			var tar = $(this);
			$.extend(obj, defaults, options);

			var name = tar.attr("name") || "",
				id = tar.attr("id") || $.su.randomId("button"),
				value = tar.val() || "",
				cls = tar.attr("class") || "";

			//运行
			var input = null;
			if (tar.is("input")){	//这边有个id重复的bug，跟input的id重复！
				var btn = $("<button id=\"button-"+id+"\" name=\"button-"+name+"\" class=\""+cls+"\" type=\"button\">"+value+"</button>");
				//console.log(btn)
				//btn.insertBefore(tar);
				//tar.addClass("hidden");
				tar.replaceWith(btn);
				input = tar;
				tar = btn;
				//obj = btn.get(0);
				$.extend(btn.get(0), defaults, options);
			};
			

			var _text = tar.html();
			var text = obj.text || _text || "";

			var inHTML = 		"";

			if (obj.iconCls !== ""){
				inHTML +=		"<span class=\"icon button-icon "+obj.iconCls+"\"></span>";
			};
				inHTML +=		"<span class=\"text button-text\">"+text+"</span>";
		
			tar.empty().addClass("button-button "+obj.btnCls).append($(inHTML));

				inHTML = 	"<div class=\"container widget-container button-container "+obj.cls+"\">";

			if (obj.fieldLabel !== null){
				inHTML +=		"<div class=\"widget-fieldlabel-wrap "+obj.labelCls+"\">";
				inHTML +=			"<label class=\"widget-fieldlabel text-fieldlabel\">"+obj.fieldLabel+"</label>";
				if (obj.fieldLabel !== ""){
					inHTML +=		"<span class=\"widget-separator\">"+obj.separator+"</span>";
				};
				inHTML +=		"</div>";			
			};

				inHTML +=		"<div class=\"widget-wrap-outer button-wrap-outer\">"
				inHTML +=			"<div class=\"widget-wrap button-wrap\">";
				inHTML +=			"</div>";

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
			var wrap = container.find("div.button-wrap");
			wrap.append(tar);

			if (input){
				wrap.append(input.addClass("hidden"));
			};



			/*if (obj.loading){
				inHTML = 	"<div class=\"loading-info-container\">";
				inHTML +=		"<span class=\"icon\"></span>";
				inHTML +=		"<span class=\"text\"></span>";
				inHTML +=	"</div>";

				container.append($(inHTML));
			};*/

			if ($.type(obj.handler) === "function"){
				tar.bind("click", obj.handler);
			};

			if (obj.enableToggle === true){
				var prs =  obj.pressed;
				tar.attr("pressed", prs);

				if (prs.toString() === "true"){
					tar.addClass(obj.pressedCls)
				};

				tar.bind("click", function(e){
					var btn = $(this),
						//_prs = btn.attr("pressed"),
						_prs = this.pressed;

					if (_prs.toString() === "false"){
						btn.addClass(this.pressedCls);
						_prs = true;
					}else{
						btn.removeClass(this.pressedCls);
						_prs = false;
					};

					btn.attr("pressed", _prs);
					this.pressed = _prs;

					if ($.type(this.toggleHander) === "function"){
						this.toggleHander.call(this, _prs);
					}
					
				})
			};

			obj.isButton = true;
		});
		
		//var container = me.button("getContainer");
		var container = me.closest("div.button-container");
		container.delegate("button.button-button", "mousedown", function(e){
			e.stopPropagation();
			container.addClass("clicked");
		}).delegate("button.button-button", "click", function(e){
			e.stopPropagation();
			e.preventDefault();
		});

		me.button("setTips", options.tips);

		return me;
	},
	setText: function(me, text){
		var me = me || this,
			container = me.closest("div.button-container"),
			btnText = container.find("button.button-button span.button-text"),
			text = text[1];
		if (text){
			me.text = text;
			btnText.html(text);
		};

		return me;
	},
	setValue: function(me, value){
		var me = me || this,
			value = value[1] || "";
		//console.log(me.get(0).nodeType, me.get(0).id);
		me.val(value)
		return me;
	},
	disable: function(me){
		var me = me || this,
			container = me.closest("div.button-container");

		container.addClass("disabled");
		container.find("button.button-button").prop("disabled", true);

		return me;
	},
	enable: function(me){
		var me = me || this,
			container = me.closest("div.button-container");

		container.removeClass("disabled");
		container.find("button.button-button").prop("disabled", false);

		return me;
	},
	getContainer: function(me) {
		var me = me || this,
			container = me.closest("div.button-container");
		return container;
	}
});

})(jQuery);