
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
  var owners = {};
	var canvas = document.getElementById("board");

	canvas.width = {{.width}};
	canvas.height = {{.height}};
	var ctx = canvas.getContext("2d");

	var targetId = function(name, x, y) {
	  return 'target_' + name + '_' + x + '-' + y;
	};
	var createTarget = function(name, precision, x, y) {
		var clientX = parseInt(parseInt(x) * (canvas.clientWidth / canvas.width));
		var clientY = parseInt(parseInt(y)* (canvas.clientHeight / canvas.height));
		var target = $('<span id="' + targetId(name, x, y) + '" class="target">' + precision + '</span>');
		target.css('color', name.colorize());
		target.css('top', '' + clientY + 'px');
		target.css('left', '' + clientX + 'px');
		$('body').append(target);
	};
	var removeTarget = function(name, x, y) {
		var clientX = parseInt(parseInt(x) * (canvas.clientWidth / canvas.width));
		var clientY = parseInt(parseInt(y)* (canvas.clientHeight / canvas.height));
		$('#' + targetId(name, x, y)).remove();
	};
	
	var socket = new WebSocket("ws://" + window.location.host + "/ws/view");
	socket.onerror = function(ev) {
	  console.log(ev);
	};
	socket.onopen = function(ev) {
	  console.log("Socket opened");
		var mold = null;
		var precision = 5;

    $('body').bind('keypress', function(ev) {
		  if (ev.charCode >= 48 && ev.charCode <= 57) {
			  if (ev.charCode == 48) {
				  precision = 100;
				} else {
				  precision = 1 + 10 * (ev.charCode - 49);
				}
			}
		});
		canvas.onmousedown = function(ev) {
			var canvasX = parseInt(ev.clientX * (canvas.width / canvas.clientWidth));
			var canvasY = parseInt(ev.clientY * (canvas.height / canvas.clientHeight));
			var pos = '' + canvasX + '-' + canvasY;
			var owner = owners[pos];
			if (owner != null) {
				mold = owner;
				$('#state').text(owner);
			}
		};
		canvas.onmouseup = function(ev) {
		  if (mold != null) {
				var canvasX = parseInt(ev.clientX * (canvas.width / canvas.clientWidth));
				var canvasY = parseInt(ev.clientY * (canvas.height / canvas.clientHeight));
				var pos = '' + canvasX + '-' + canvasY;
				var owner = owners[pos];
				if (owner == mold) {
				  socket.send(JSON.stringify({
					  Op: 'clearTargets',
						Name: mold,
					}));
				} else {
					socket.send(JSON.stringify({
						Op: 'createTarget',
						Target: [canvasX, canvasY],
						Name: mold,
						Precision: precision,
					}));
				}
			}
		}
	};
	socket.onclose = function(ev) {
		console.log("Socket closed");
	};
	socket.onmessage = function(ev) {
		var obj = JSON.parse(ev.data);
		if (obj.Molds) {
		  for (var mold in obj.Molds) {
			  for (var i = 0; i < obj.Molds[mold].Bits.length; i++) {
				  var p = obj.Molds[mold].Bits[i];
					if (p != null) {
						owners['' + p.X + '-' + p.Y] = mold;
						set(ctx, p.X, p.Y, mold.colorize());
					}
				}
			}
			for (var mold in obj.Molds) {
			  for (var targ in obj.Molds[mold].Targets) {
				  var xy = targ.split("-");
				  createTarget(mold, obj.Molds[mold].Targets[targ], xy[0], xy[1]);
				}
			}
		} else {
			for (var pos in obj.Created) {
			  owners[pos] = obj.Created[pos];
				var xy = pos.split("-");
				set(ctx, parseInt(xy[0]), parseInt(xy[1]), obj.Created[pos].colorize());
			}
			for (var pos in obj.Removed) {
			  delete(owners[pos]);
				var xy = pos.split("-");
				unset(ctx, parseInt(xy[0]), parseInt(xy[1]));
			}
			for (var mold in obj.CreatedTargets) {
			  for (var targ in obj.CreatedTargets[mold]) {
					var xy = targ.split("-");
					createTarget(mold, obj.CreatedTargets[mold][targ], xy[0], xy[1]);
				}
			}
			for (var mold in obj.RemovedTargets) {
			  for (var targ in obj.RemovedTargets[mold]) {
					var xy = targ.split("-");
					removeTarget(mold, xy[0], xy[1]);
				}
			}
		}
	};
});

