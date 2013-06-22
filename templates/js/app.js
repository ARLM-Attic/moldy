
String.prototype.colorize = function() {
	var hash = murmurhash3_32_gc(this, 1);
	rval = "#";
	rval += (hash & 0xff).toString(16);
	if (rval.length < 3) rval += "0"; 
	rval += ((hash & 0xff00) >> 8).toString(16);
	if (rval.length < 5) rval += "0";
	rval += ((hash & 0xff0000) >> 16).toString(16);
	if (rval.length < 7) rval += "0";
	return rval;
};

function set(ctx, x, y, col) {
	ctx.fillStyle = col;
	ctx.fillRect(x, y, 1, 1);
};

function unset(ctx, x, y) {
	ctx.fillStyle = "white";
	ctx.fillRect(x, y, 1, 1);
};

$(function() {
	var canvas = document.getElementById("board");
	canvas.width = {{.width}};
	canvas.height = {{.height}};
	var ctx = canvas.getContext("2d");
	
	var socket = new WebSocket("ws://" + window.location.host + "/ws/view");
	socket.onerror = function(ev) {
	  console.log(ev);
	};
	socket.onopen = function(ev) {
	  console.log("Socket opened");
	};
	socket.onclose = function(ev) {
	  console.log("Socket closed");
	};
	socket.onmessage = function(ev) {
		var obj = JSON.parse(ev.data);
		if (obj.Created && obj.Removed) {
			for (var pos in obj.Created) {
				var xy = pos.split("-");
				set(ctx, parseInt(xy[0]), parseInt(xy[1]), obj.Created[pos].colorize());
			}
			for (var pos in obj.Removed) {
				var xy = pos.split("-");
				unset(ctx, parseInt(xy[0]), parseInt(xy[1]));
			}
		} else {
		  for (var mold in obj.Molds) {
			  for (var pos in obj.Molds[mold].Bits) {
					var xy = pos.split("-");
					set(ctx, parseInt(xy[0]), parseInt(xy[1]), mold.colorize());
				}
			}
		}
	};
});

