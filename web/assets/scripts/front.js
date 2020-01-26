"use strict";
$(document).ready(function () {
    $(document.forms["login-form"]).submit(function (event) {
        event.preventDefault();
        $.ajax(
            $(this).attr("action"),
            {
                type: $(this).attr("method"),
                data: $(this).serialize(),
                contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
                processData: false,
                success: function () {
                    window.location.reload();
                },
                error: function (jqXHR) {
                    document.forms["login-form"].innerHTML = jqXHR.responseText
                }
            },
        );
    });
    $(document.forms["signup-form"]).submit(function (event) {
        event.preventDefault();
        $.ajax(
            $(this).attr("action"),
            {
                type: $(this).attr("method"),
                data: $(this).serialize(),
                contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
                processData: false,
                success: function () {
                    window.location.reload();
                },
                error: function (jqXHR) {
                    document.forms["signup-form"].innerHTML = jqXHR.responseText
                }
            },
        );
    });
});

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