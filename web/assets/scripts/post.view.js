"use strict";
$(document).ready(function () {
    populateCommentBoard(
        window.location.href + "/comment-board",
        "#comments-list",
        1, 25, "");
    $("#add-comment-form").submit(function (event) {
        event.preventDefault();
        $.ajax(
            window.location.href + "/add-comment",
            {
                type: "POST",
                data: JSON.stringify({
                    Comment: $("#comment-ta").val(),
                    CSRF: $("#_csrf").val()
                }),
                contentType: "application/json",
                success: function updatePageDisplay(data, textStatus, jqXHR) {
                    populateCommentBoard(
                        window.location.href + "/comment-board",
                        "#comments-list",
                        1, 25, "");
                },
                error: function (jqXHR, textStatus, errorThrown) {
                    $(this).html(jqXHR.responseText);
                }
            }
        )
    });
});

function populateCommentBoard(url, container, page, perPage, sorting) {
    $.ajax(
        url,
        {
            type: "POST",
            data: JSON.stringify({page: page, perPage: perPage, sorting: sorting}),
            contentType: "application/json",
            success: function updatePageDisplay(data, textStatus, jqXHR) {
                $(container).html(data.toString());
            },
            error: function () {
                $(container).html("<h1> No Comments are available.</h1>");
            }
        }
    )
}

function simpleTextFormSubmit(name, event) {
    event.preventDefault();
    // var formData = new FormData(this);
    let form = $('#' + name);
    $.ajax(
        form.attr("action"),
        {
            type: form.attr("method"),
            data: form.serialize(),
            contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
            processData: false,
            success: function (data, textStatus, jqXHR) {
                window.location.reload();
            },
            error: function (jqXHR, textStatus, errorThrown) {
                document.forms[name].innerHTML = jqXHR.responseText
            }
        },
    );
}
;