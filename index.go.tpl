<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css" integrity="sha384-B0vP5xmATw1+K9KRQjQERJvTumQW0nPEzvF6L/Z6nronJ3oUOFUFpCjEUQouq2+l" crossorigin="anonymous">
<script src="https://cdn.jsdelivr.net/npm/jquery@3.5.1/dist/jquery.slim.min.js" integrity="sha384-DfXdz2htPH0lsSSs5nCTpuj/zy4C+OGpamoFVy38MVBnE+IbbVYUew+OrCXaRkfj" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/js/bootstrap.bundle.min.js" integrity="sha384-LCPyFKQyML7mqtS+4XytolfqyqSlcbB3bvDuH9vX2sdQMxRonb/M3b9EmhCNNNrV" crossorigin="anonymous"></script>
</head>
<body>
{{range .Videolist }}
    <div class="col-sm-6 col-md-2">
       <div class="card" style="width: 18rem;">
         <img src=" {{.ImgUrl}}" class="card-img-top" alt="...">
         <div class="card-body">
           <p class="card-text">{{.Title}}</p>
           <a href="{{.Path}}" class="btn btn-primary">跳转</a>
         </div>
       </div>
    </div>
{{ end }}
</body>
</html>
