"use strict";
$(document).ready(function () {
    populatePostList(
        "/home-feed-posts",
        ".post-list",
        1, 25, ""
    );
    var timerID;
    $("#search-input").on("input", function search(event) {
        var value = $(event.currentTarget).val();
        if ($(event.currentTarget).data("lastval") != value) {

            $(event.currentTarget).data("lastval", value);

            clearTimeout(timerID);

            timerID = setTimeout(function () {
                $.ajax(
                    "http://localhost:8080/search",
                    {
                        type: "GET",
                        data: {
                            pattern: $(event.currentTarget).val(),
                            limit: 5,
                            offset: 0,
                            sort: $("input[name='gender']:checked").val() + "_" + $("input[name='sort-order']:checked").val()
                        },
                        dataType: "json",
                        success: function updatePageDisplay(data, textStatus, jqXHR) {
                            $("#search-results").html("");
                            var results = data.data;
                            console.log(results);
                            if (Array.isArray(results.Posts)) {
                                for (let i = 0; i < results.Posts.length; i++) {
                                    let p = results.Posts[i];
                                    $("#search-results").append(`<div class="container mt-3">
                                        <div class="media border p-3">
                                            <img src="../assets/img/p.png" class="mr-3 mt-3 rounded-circle"
                                                 style="width:60px;">
                                            <div class="media-body">
                                                <h4>` + p.title + `<span style="margin: 0 0.5%; font-size: small; "><a
                                                                href="#">@` + p.originChannel + `</a></span>
                                                </h4>
                                                <p>` + p.description + `</p>
                                            </div>
                                        </div>
                                    </div>`)
                                }
                            }
                            if (Array.isArray(results.Releases)) {
                                for (let i = 0; i < results.Releases.length; i++) {
                                    let p = results.Releases[i];
                                    $("#search-results").append(`<div class="container mt-3">
                                        <div class="media border p-3">
                                            <img src="../assets/img/r.png" class="mr-3 mt-3 rounded-circle"
                                                 style="width:60px;">
                                            <div class="media-body">
                                                <h4>` + p.metadata.title + `<span style="margin: 0 0.5%; font-size: small; "><a
                                                                href="#">` + p.ownerChannel + `</a></span>
                                                </h4>
                                                <p>` + p.metadata.description + `</p>
                                            </div>
                                        </div>
                                    </div>`)
                                }
                            }
                            if (Array.isArray(results.Channels)) {
                                for (let i = 0; i < results.Channels.length; i++) {
                                    let p = results.Channels[i];
                                    $("#search-results").append(`<div class="container mt-3">
                                        <div class="container mt-3">
                                        <div class="media border p-3">
                                            <img src="../assets/img/c.png" class="mr-3 mt-3 rounded-circle"
                                                 style="width:60px;">
                                            <div class="media-body">
                                                <h4>` + p.name + `<span style="margin: 0 0.5%; font-size: small; "><a
                                                                href="#">@` + p.channelUsername + `</a></span>
                                                </h4>
                                                <p>` + p.description + `</p>
                                            </div>
                                        </div>
                                    </div>`)
                                }
                            }
                            if (Array.isArray(results.Users)) {
                                for (let i = 0; i < results.Users.length; i++) {
                                    let p = results.Users[i];
                                    console.log(p);
                                    $("#search-results").append(`<div class="container mt-3">
                                        <div class="media border p-3">
                                            <img src="` + p.pictureURL + `" class="mr-3 mt-3 rounded-circle"
                                                 style="width:60px;">
                                            <div class="media-body">
                                                <h4>` + p.firstName + ' ' + p.middleName + ' ' + p.lastName + `
                                                <span style="margin: 0 0.5%; font-size: small; "><a
                                                                href="#">@` + p.username + `</a></span></h4>
                                                <p>` + p.bio + `</p>
                                            </div>
                                        </div>
                                    </div>`
                                    )
                                }
                            }
                        },
                        error: function (jqXHR, textStatus, errorThrown) {
                            console.log(jqXHR.responseText);
                            console.log(textStatus);
                            console.log(errorThrown);
                            $("#search-results").html("<h2>Server Error.</h2>")
                        }
                    }
                )
            }, 500);
        }
    });
});

function populatePostList(url, container, page, perPage, sorting) {
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
                $(container).html("<h1> No posts are available.</h1>");
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