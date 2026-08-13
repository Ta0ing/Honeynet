$(document).ready(function() {

    var select_toggle = false;

    // Tool tips & popovers
    $('.tt_bind').tooltip();
    $('.po_bind').popover();

    // Date selectors
    $('#range_start').datetimepicker();
    $('#range_end').datetimepicker();

    // Create generic time range selector
    $('#timerange').change(function() {
        set_time_range_selection();
    });

    // Create bindings for select/deselect all
    $('.select_all').click(function() {
        var forname = $(this).data('for-name');
        if (select_toggle) {
            $('input[name=' + forname + ']').prop('checked', false);
            select_toggle = false;
            $(this).removeClass('icon-minus-sign').addClass('icon-plus-sign');
            $(this).attr('title', 'Select All');
            $('input[name=' + forname + ']').trigger('change');
        } else {
            $('input[name=' + forname + ']').prop('checked', true);
            select_toggle = true;
            $(this).removeClass('icon-plus-sign').addClass('icon-minus-sign');
            $(this).attr('title', 'De-select All');
            $('input[name=' + forname + ']').trigger('change');
        }
    });

    var resize_admin_navbar = function() {
        if ($('.admin-leftbar').length) {
            var height = $(window).height() - 268;
            $('.admin-leftbar').css('height', height+"px");
            $('.admin-leftbar .well:last').css('height', height+"px");
        }
    }

    resize_admin_navbar();
    $(window).resize(function() {
        resize_admin_navbar();
    });

});

function human_readable_lifetime(lifetime)
{
    letter = lifetime.substring(lifetime.length-1);
    num = lifetime.substring(0, lifetime.length-1);
    if (letter == 'H' || letter == 'h') {
        return num + ' Hours';
    } else if (letter == 'D' || letter == 'd') {
        return num + ' Days';
    } else if (letter == 'W' || letter == 'w') {
        return num + ' Weeks';
    }

    return lifetime;
}

function human_readable_toptype(toptype)
{
    switch (toptype)
    {
        case 'srcip':
            return "source IP";

        case 'dstip':
            return "destination IP";

        case 'srcport':
            return "source port";

        case 'dstport':
            return "destination port";
    }
}

function human_readable_timerange(start, end)
{
    if (start == '-2 hours' && end == '-1 second') {
        return "last 2 hours";
    } else if (start == '-4 hours' && end == '-1 second') {
        return "last 4 hours";
    } else if (start == '-6 hours' && end == '-1 second') {
        return "last 6 hours";
    } else if (start == '-12 hours' && end == '-1 second') {
        return "last 12 hours";
    } else if (start == '-24 hours' && end == '-1 second') {
        return "last 24 hours";
    } else if (start == '-2 days' && end == '-1 second') {
        return "last 2 days";
    } else if (start == '-1 week' && end == '-1 second') {
        return "last week";
    } else if (start == '-1 month' && end == '-1 second') {
        return "last month";
    } else if (end == '-1 second') {
        return start.replace('-', '') + " ago";
    } else {
        return make_datepicker_timestamp(start) + " to " + make_datepicker_timestamp(end);
    }
}

function human_readable_size(size, bytes) {
    var gb = 1073741824; // bytes
    var mb = 1048576;
    var kb = 1024;
    
    if (bytes == undefined) {
        var append = 'B';
    } else {
        var append = '';
    }
    
    if (size > gb) {
        return (size / gb).toFixed(2) + " Gi" + append;
    } else if (size > mb) {
        return (size / mb).toFixed(2) + " Mi" + append;
    } else if (size > kb) {
        return (size / kb).toFixed(2) + " K" + append;
    } else {
         return Math.round(size) + " " + append;
    }
}

function loading(obj) {
    var loader = base_url + 'media/images/ajax-loader.gif';
    obj.html('<img src="' + loader + '" />');
}

function loading_small(obj) {
    var loader = base_url + 'media/images/ajax-loader.gif';
    obj.html('<img width="16" height="16" src="' + loader + '" />');
}

function make_nfdump_timestamp(time) {
    var date = time.split(' ');
    var front = date[0].split('/');
    var new_time = front[2] + '/' + front[0] + '/' + front[1] + '.' + date[1] + ':59';
    return new_time;
}

function make_datepicker_timestamp(time) {
    var date = time.split('.');
    var front = date[0].split('/');
    var back = date[1].split(':');
    var new_time = front[1] + '/' + front[2] + '/' + front[0] + ' ' + back[0] + ':' + back[1];
    return new_time;
}

function set_time_range(start, end)
{
    var tr = $('#timerange');
    if (start == '-2 hours' && end == '-1 second') {
        tr.val('2h');
    } else if (start == '-4 hours' && end == '-1 second') {
        tr.val('4h');
    } else if (start == '-6 hours' && end == '-1 second') {
        tr.val('6h');
    } else if (start == '-12 hours' && end == '-1 second') {
        tr.val('12h');
    } else if (start == '-24 hours' && end == '-1 second') {
        tr.val('24h');
    } else if (start == '-2 days' && end == '-1 second') {
        tr.val('2d');
    } else if (start == '-1 week' && end == '-1 second') {
        tr.val('1w');
    } else if (start == '-1 month' && end == '-1 second') {
        tr.val('1m');
    } else if (end == '-1 second') {
        tr.val('elapsed').trigger('change');
        $('#elapsed_start').val(start.replace('-', ''));
        $('#elapsed_end').val('1 second');
    } else {
        tr.val('range').trigger('change');
        $('#range_start').val(make_datepicker_timestamp(start));
        $('#range_end').val(make_datepicker_timestamp(end));
    }
    set_time_range_selection();
}

function set_time_range_selection()
{
    if ($('#timerange').val() == 'elapsed') {
        $('#elapsed_inputs').show();
        $('#range_inputs').hide();
    } else if ($('#timerange').val() == 'range') { 
        $('#elapsed_inputs').hide();
        $('#range_inputs').show();
    } else {
        $('#elapsed_inputs').hide();
        $('#range_inputs').hide();
    }
}

function get_correct_datetimes()
{
    var select = $('#timerange').val();
    var dates = { 'start': '', 'end': '-1 second' }
    if (select == '2h') {
        dates.start = '-2 hours';
    } else if (select == '4h') {
        dates.start = '-4 hours';
    } else if (select == '6h') {
        dates.start = '-6 hours';
    } else if (select == '12h') {
        dates.start = '-12 hours';
    } else if (select == '24h') {
        dates.start = '-24 hours';
    } else if (select == '2d') {
        dates.start = '-2 days';
    } else if (select == '1w') {
        dates.start = '-1 week';
    } else if (select == '1m') {
        dates.start = '-1 month';
    } else if (select == 'range') {
        dates.start = make_nfdump_timestamp($('#range_start').val());
        dates.end = make_nfdump_timestamp($('#range_end').val());
    } else if (select == 'elapsed') {
        dates.start = '-' + $('#elapsed_start').val();
        dates.end = '-' + $('#elapsed_end').val();
    }
    return dates;
}

function flowtype(type)
{
    switch (type) {
        case 'netflow':
            return "NetFlow";
        case 'sflow':
            return "sFlow";
        default:
            return;
    }
}

/* Helpers for graph visualizations */

// Change the bg full size on pages that use graph popups
function resize_bg_full() {
    var width = $(window).width();
    var height = $(window).height();
    $('#bg-full').css('width', width)
                 .css('height', height)
                 .css('position', 'fixed')
                 .css('top', 0)
                 .css('left', 0)
                 .css('opacity', 0.85);
}
