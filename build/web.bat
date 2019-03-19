:: Copyright 2017 by caixw, All rights reserved.
:: Use of this source code is governed by a MIT
:: license that can be found in the LICENSE file.


set wd=%~dp0\..\cmd\web

set mainPath=github.com\issue9\web

set varsPath=%mainPath%\internal\cmd\version

set builddate=%date:~0,4%%date:~5,2%%date:~8,2% 

%GOROOT%\bin\go build -o %wd%\web.exe -ldflags "-X %varsPath%.buildDate=%builddate%" -v %mainPath%\cmd\web
