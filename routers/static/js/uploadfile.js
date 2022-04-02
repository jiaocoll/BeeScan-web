$(function (){
    var f = document.getElementById("upflag").value;
    if (f != ""){
        alert(f)
    }
})
$(function (){
    var a = document.getElementById("loginmsg").value;
    if (a != ""){
        alert(a)
    }
})
function Uploadfile() {
    $("#pocfile").click();
    $("#pocfile").change(function (){
        $("#UpLodeFile").submit();
    })
}

