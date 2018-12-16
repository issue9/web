:: Copyright 2017 by caixw, All rights reserved.
:: Use of this source code is governed by a MIT
:: license that can be found in the LICENSE file.


:: 代码的主目录，变量赋值时，等号两边不能有空格。
set wd=%~dp0\..\cmd\web

:: 程序所在的目录
set mainPath=github.com\issue9\web

:: 需要修改变量的名名，若为 main，则指接使用 main，而不是全地址
set varsPath=%mainPath%\internal\cmd\version

:: 当前日期，格式为 YYYYMMDD
set builddate=%date:~0,4%%date:~5,2%%date:~8,2% 

echo 开始编译
%GOROOT%\bin\go build -o %wd%\web.exe -ldflags "-X %varsPath%.buildDate=%builddate%" -v %mainPath%\cmd\web
