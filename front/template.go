package front

var indexTemplateSource = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css"
          integrity="sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh" crossorigin="anonymous">
    <style>
        .img-wrapper {
            max-width: 600px;
            max-height: 600px;
            position: relative;
        }

        .img-wrapper img {
            max-width: 600px;
            max-height: 600px;
        }

        .clearfix:after {
            content: '';
            display: block;
            clear: both;
        }

        .previews {
            display: flex;
            flex-flow: row;
            justify-content: space-between;
            align-content: flex-start;
            flex-wrap: wrap;
            margin-bottom: 20px;
        }

        .preview {
            width: 102px;
            height: 120px;
            border: 1px solid #e4e4e4;
            text-align: center;
        }

        .preview_current {
            /*box-shadow: 0 0 10px black;*/
            outline: 2px dashed black;
        }

        .preview__link {
            display: inline-block;
            width: 100px;
            height: 100px;
        }

        .preview__img-wrapper {
            display: table-cell;
            vertical-align: middle;
            height: 100px;
            width: 100px;
        }

        .preview img {
            display: inline-block;
            max-height: 100px;
            max-width: 100px;
        }

        .preview__caption {
            font-size: 10px;
            color: grey;
            height: 20px;
            line-height: 20px;
            overflow: hidden;
        }

        #canvas {
            cursor: crosshair;
            position: absolute;
            top: 0;
            left: 0;
            z-index: 10;
        }

        .area-tip {
            font-size: 12px;
            color: #999;
        }
    </style>
    <script>
        var filename;
        var rect;
        var label;

        function restoreData() {
            filename = document.getElementsByName("filename")[0].value;

            var prevFilename = sessionStorage.getItem("filename");
            if (prevFilename === filename) {
                // we already edited this image. Let's restore it's data
                var storedLabel = sessionStorage.getItem('label');
                if (typeof storedLabel !== 'undefined' && storedLabel !== null) {
                    document.getElementsByName('label')[0].value = storedLabel;
                }
                var storedRect = sessionStorage.getItem('rect');
                if (typeof storedRect !== 'undefined' && storedRect !== null) {
                    try {
                        storedRect = JSON.parse(storedRect);
                        rect = storedRect;
                    } catch (ex) {
                        console.error('Error parsing stored rect', ex)
                    }
                }
            }

            if (typeof rect === 'undefined' || rect === null) {
                // it's a new image. Initialize data
                rect = {
                    left: 0,
                    top: 0,
                    right: 0,
                    bottom: 0
                };
            }
            // update form
            storeData();
        }

        function storeData() {
            var label = document.getElementsByName('label')[0].value;

            sessionStorage.setItem('filename', filename);
            sessionStorage.setItem('rect', JSON.stringify(rect));
            sessionStorage.setItem('label', label);

            var l = Math.min(rect.left, rect.right);
            var r = Math.max(rect.left, rect.right);
            var t = Math.min(rect.top, rect.bottom);
            var b = Math.max(rect.top, rect.bottom);

            // our image can be scaled. Let's find scale coefficients
            // (in general width and height are scale by the same coeff, but let's play it safe)
            var img = document.getElementById('img');
            var scw = img.naturalWidth / img.clientWidth;
            var sch = img.naturalHeight / img.clientHeight;

            l = Math.round(l * scw);
            r = Math.round(r * scw);
            t = Math.round(t * sch);
            b = Math.round(b * sch);

            document.getElementsByName('width')[0].value = img.naturalWidth;
            document.getElementsByName('height')[0].value = img.naturalHeight;

            document.getElementsByName('left')[0].value = l;
            document.getElementsByName('top')[0].value = t;
            document.getElementsByName('right')[0].value = r;
            document.getElementsByName('bottom')[0].value = b;

            document.getElementById('area-tip').innerHTML = ` + "`" + `Selected area left: ${l} top: ${t} right: ${r} bottom: ${b}` + "`" + `;

            var ok = true;
            if (l < r && t < b) {
                document.getElementById('area-tip').classList.remove('area-tip_warn');
            } else {
                document.getElementById('area-tip').innerHTML += ` + "`" + `<br/><span class="badge badge-danger">SELECTION HAS ZERO SIZE!</span>` + "`" + `;
                document.getElementById('area-tip').classList.add('area-tip_warn');
                ok = false;
            }

            if (typeof label !== 'undefined' && label !== null && label.trim() !== '') {
                // empty label
                document.getElementsByName('label')[0].classList.remove("is-invalid");
            } else {
                document.getElementsByName('label')[0].classList.add("is-invalid");
                ok = false;
            }

            if (ok) {
                document.getElementById('submit').disabled = false;
            } else {
                document.getElementById('submit').disabled = true;
            }
        }

        function draw() {
            var c = document.getElementById('canvas');
            var ctx = c.getContext('2d');

            ctx.clearRect(0, 0, c.width, c.height);
            ctx.fillStyle = "rgba(0, 255, 0, 0.3)";
            ctx.fillRect(0, 0, c.width, c.height);

            ctx.strokeStyle = 'lime';
            ctx.lineWidth = '1px';
            ctx.clearRect(rect.left, rect.top, rect.right - rect.left, rect.bottom - rect.top);
            ctx.strokeRect(rect.left, rect.top, rect.right - rect.left, rect.bottom - rect.top);
        }

        function resizeCanvas() {
            var c = document.getElementById('canvas');
            var img = document.getElementById('img');

            console.log(img.clientWidth, img.clientHeight, img.naturalWidth, img.naturalHeight);

            c.width = img.clientWidth;
            c.height = img.clientHeight;
        }

        function onLoad() {
            restoreData();

            var isDrawing = false;

            var c = document.getElementById('canvas');
            c.addEventListener('mousedown', function (ev) {
                rect.left = ev.offsetX;
                rect.top = ev.offsetY;
                rect.right = rect.left;
                rect.bottom = rect.top;
                isDrawing = true;
                requestAnimationFrame(draw);
            });
            c.addEventListener('mousemove', function (ev) {
                if (isDrawing) {
                    rect.right = ev.offsetX;
                    rect.bottom = ev.offsetY;
                    requestAnimationFrame(draw);
                }
            });
            c.addEventListener('mouseup', function (ev) {
                if (isDrawing) {
                    isDrawing = false;
                    rect.right = ev.offsetX;
                    rect.bottom = ev.offsetY;
                    storeData();
                    requestAnimationFrame(draw);
                }
            });
            c.addEventListener('mouseout', function (ev) {
                isDrawing = false;
                storeData();
                requestAnimationFrame(draw);
            });

            document.getElementsByName('label')[0].addEventListener('change', storeData);

            resizeCanvas();
            requestAnimationFrame(draw);
        }
    </script>
</head>
<body onload="onLoad()">
<div class="container">
    {{if .Previews}}
        <p>Showing {{.PreviewLeft}} - {{.PreviewRight}} of total {{.TotalFiles}} images</p>
        <div class="previews">
            {{range .Previews}}
                <div class="preview {{if eq $.Filename .}}preview_current{{end}}">
                    <a href="?filename={{.}}" class="preview__link">
                        <div class="preview__img-wrapper">
                            <img src="/img/{{.}}" alt="{{.}}">
                        </div>
                        <div class="preview__caption">{{.}}</div>
                    </a>
                </div>
            {{end}}
        </div>
    {{else}}
        <p class="alert alert-warning" role="alert">No files found for previews.</p>
    {{end}}
    {{if .Errors}}
        <div>
            {{range .Errors}}
                <p class="alert alert-danger" role="alert">{{.}}</p>
            {{end}}
        </div>
    {{end}}
    {{if .Filename}}
        <form action="/process" method="POST">
            <h3>{{.Filename}}</h3>
            <div class="form-group">
                <div id="wrapper" class="img-wrapper clearfix">
                    <img id="img" src="/img/{{.Filename}}" onload="resizeCanvas()"/>
                    <canvas id="canvas"></canvas>
                </div>
                <p id="area-tip" class="area-tip"></p>
            </div>
            <div class="form-group">
                <label for="label-input">Area label</label>
                <input id="label-input" type="text" class="form-control" name="label"
                       placeholder="enter area label here"
                       required/>
            </div>
            <div class="form-group">
                <input type="hidden" name="filename" value="{{.Filename}}"/>
                <input type="hidden" name="width" value="0"/>
                <input type="hidden" name="height" value="0"/>
                <input type="hidden" name="left" value="0"/>
                <input type="hidden" name="top" value="0"/>
                <input type="hidden" name="right" value="0"/>
                <input type="hidden" name="bottom" value="0"/>
            </div>
            <div class="form-group">
                <input id="submit" type="submit" class="btn btn-primary" value="save"/>
            </div>
        </form>
    {{else}}
        <div class="alert alert-danger">No image found to edit. Images directory is empty?</div>
    {{end}}
</div>
</body>
</html>`
