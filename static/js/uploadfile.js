$(function (){
    var f = document.getElementById("upflag").value;
    if (f != ""){
        alert(f)
    }
})
function Uploadfile() {
    $("#pocfile").click();
    $("#pocfile").change(function (){
        $("#UpLodeFile").submit();
    })
}

