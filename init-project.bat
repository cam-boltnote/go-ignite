@echo off
setlocal enabledelayedexpansion

:: Get project name from user
set /p PROJECT_NAME="Enter your project name (e.g., my-activity-api): "

:: Get the current directory (go-ignite root) and its parent
set "CURRENT_DIR=%~dp0"
set "CURRENT_DIR=%CURRENT_DIR:~0,-1%"
set "PARENT_DIR=%CURRENT_DIR%\.."

:: Create new directory in the parent directory of go-ignite
set "NEW_PROJECT_DIR=%PARENT_DIR%\%PROJECT_NAME%"

:: Create new directory
mkdir "%NEW_PROJECT_DIR%"

:: Copy necessary files and directories
echo Copying project files...
xcopy /E /I /Y "%CURRENT_DIR%\cmd" "%NEW_PROJECT_DIR%\cmd"
xcopy /E /I /Y "%CURRENT_DIR%\internal" "%NEW_PROJECT_DIR%\internal"
xcopy /E /I /Y "%CURRENT_DIR%\docs" "%NEW_PROJECT_DIR%\docs"
xcopy /E /I /Y "%CURRENT_DIR%\tools" "%NEW_PROJECT_DIR%\tools"
copy /Y "%CURRENT_DIR%\.env.example" "%NEW_PROJECT_DIR%\.env.example"
copy /Y "%CURRENT_DIR%\.air.toml" "%NEW_PROJECT_DIR%\.air.toml"
copy /Y "%CURRENT_DIR%\.gitignore" "%NEW_PROJECT_DIR%\.gitignore"
copy /Y "%CURRENT_DIR%\README.md" "%NEW_PROJECT_DIR%\README.md"

:: Create basic .env file
echo Creating basic .env file...
(
echo # Basic Configuration
echo PORT=8080
echo GIN_MODE=debug
echo LOG_LEVEL=info
echo.
echo # Proxy Configuration
echo TRUSTED_PROXIES=0.0.0.0/0
echo.
echo # Database Configuration
echo INIT_DB=false
echo AUTO_MIGRATE=false
echo DB_USER=root
echo DB_PASSWORD=root
echo DB_HOST=localhost
echo DB_PORT=3306
echo DB_NAME=%PROJECT_NAME%
echo.
echo # Encryption Configuration
echo ENCRYPTION_KEY=your_32_byte_encryption_key_here
echo.
echo # Email Configuration
echo INIT_SMTP=false
echo SMTP_HOST=smtp.example.com
echo SMTP_PORT=587
echo SMTP_USERNAME=your_smtp_username
echo SMTP_PASSWORD=your_smtp_password
echo SMTP_FROM_EMAIL=noreply@example.com
) > "%NEW_PROJECT_DIR%\.env"

:: Create new go.mod file
echo Creating new go.mod file...
cd "%NEW_PROJECT_DIR%"
go mod init %PROJECT_NAME%

:: Update import paths in all Go files
echo Updating import paths...
for /r %%f in (*.go) do (
    powershell -Command "(Get-Content '%%f') -replace 'github.com/cam-boltnote/go-ignite', '%PROJECT_NAME%' | Set-Content '%%f'"
)

:: Update go.mod dependencies
echo Updating dependencies...
go mod tidy

:: Initialize git repository
echo Initializing git repository...
git init

:: Create initial commit
git add .
git commit -m "Initial commit"

echo.
echo Project initialization complete!
echo.
echo Next steps:
echo 1. cd %NEW_PROJECT_DIR%
echo 2. Update the .env file with your specific configuration values
echo 3. Set INIT_DB=true in .env if you want to initialize the database
echo 4. Set INIT_SMTP=true in .env if you want to initialize the email sender
echo 5. Update TRUSTED_PROXIES in .env with your proxy configuration
echo 6. Run 'go mod tidy' to download dependencies
echo 7. Run 'go run cmd/main.go' to start the server
echo.
pause 