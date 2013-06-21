
function set(x, y, col) {
  unset(x, y);
  var cellWidth = $(document).width() / {{.width}};
	var cellHeight = $(document).height() / {{.height}};
  var bit = $('<div></div>');
	bit.addClass('pos' + x + '-' + y);
	bit.css('position', 'absolute');
	bit.css('left', '' + (cellWidth * x) + 'px');
	bit.css('width', '' + cellWidth + 'px');
	bit.css('height', '' + cellHeight + 'px');
	bit.css('top', '' + (cellHeight * y) + 'px');
	bit.css('background-color', col);
	$('body').append(bit);
};

function unset(x, y) {
  $('.pos' + x + '-' + y).remove();
};

$(function() {
});

