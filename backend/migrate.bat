@echo off
REM Migration helper script for Windows

set ACTION=%1
set NAME=%2
set STEPS=%3
set VERSION=%4

if "%ACTION%"=="" set ACTION=help
if "%STEPS%"=="" set STEPS=1

if "%ACTION%"=="up" goto migrate_up
if "%ACTION%"=="down" goto migrate_down
if "%ACTION%"=="create" goto migrate_create
if "%ACTION%"=="version" goto migrate_version
if "%ACTION%"=="force" goto migrate_force
if "%ACTION%"=="drop" goto migrate_drop
goto help

:migrate_up
echo Running migrations...
go run cmd/migrate.go -action=up
goto end

:migrate_down
echo Rolling back %STEPS% migration(s)...
go run cmd/migrate.go -action=down -steps=%STEPS%
goto end

:migrate_create
if "%NAME%"=="" (
    echo Error: Please specify migration name
    echo Usage: migrate.bat create migration_name
    exit /b 1
)
echo Creating migration: %NAME%
go run cmd/migrate.go -action=create -name=%NAME%
goto end

:migrate_version
echo Checking migration version...
go run cmd/migrate.go -action=version
goto end

:migrate_force
if "%VERSION%"=="" (
    echo Error: Please specify version
    echo Usage: migrate.bat force version
    exit /b 1
)
echo Forcing migration to version %VERSION%...
go run cmd/migrate.go -action=force -version=%VERSION%
goto end

:migrate_drop
echo WARNING: This will drop all tables!
set /p confirm="Are you sure? Type 'yes' to confirm: "
if not "%confirm%"=="yes" (
    echo Aborted
    exit /b 0
)
go run cmd/migrate.go -action=drop
goto end

:help
echo Migration Helper Script
echo.
echo Usage: migrate.bat [command] [args]
echo.
echo Commands:
echo   up                   - Run all pending migrations
echo   down [steps]         - Rollback migrations (default: 1 step)
echo   create ^<name^>        - Create new migration files
echo   version              - Show current migration version
echo   force ^<version^>      - Force migration version (use with caution)
echo   drop                 - Drop all tables (dangerous!)
echo   help                 - Show this help message
echo.
echo Examples:
echo   migrate.bat up
echo   migrate.bat down 2
echo   migrate.bat create add_users_table
echo   migrate.bat version
echo   migrate.bat force 1
echo.

:end

