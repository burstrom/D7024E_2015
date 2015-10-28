$(document).ready(function() {
    $("#lesen").click(function() {
        $.ajax({
            url : "helloworld.txt",
            dataType: "text",
            success : function (data) {
                $(".text").html(data);
            }
        });
    });
});