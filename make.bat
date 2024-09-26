@echo off
set time_hh=%time:~0,2%
if /i %time_hh% LSS 10 (set time_hh=0%time:~1,1%)
set BUILD_DATE=%date:~,4%-%date:~5,2%-%date:~8,2% %time_hh%:%time:~3,2%:%time:~6,2% 

for /F %%i in ('git rev-parse HEAD') do ( set COMMIT_HASH=%%i)


go build  -trimpath -ldflags "-w -s" -ldflags "-X \"main.BuildVersion=%COMMIT_HASH%\" -X \"main.BuildDate=%BUILD_DATE%\"" 