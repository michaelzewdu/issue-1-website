"use strict";
$(document).ready(function () {
    $(document.forms["login-form"]).submit(simpleTextFormSubmit);
    $(document.forms["signup-form"]).submit(simpleTextFormSubmit);
});

function simpleTextFormSubmit(event) {
    event.preventDefault();
    // document.getElementsByTagName("body")[0].hidden = true;
    // var formData = new FormData(this);
    let form;
    form = $(event.currentTarget);
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
                form.html(jqXHR.responseText);
            }
        },
    );
}