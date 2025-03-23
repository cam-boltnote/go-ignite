#!/bin/bash

# Get project name from user
read -p "Enter your project name (e.g., my-activity-api): " PROJECT_NAME

# Get the current directory (go-ignite root) and its parent
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENT_DIR="$(dirname "$CURRENT_DIR")"

# Create new directory in the parent directory of go-ignite
NEW_PROJECT_DIR="$PARENT_DIR/$PROJECT_NAME"

# Create new directory
mkdir -p "$NEW_PROJECT_DIR"

# Copy necessary files and directories
echo "Copying project files..."
cp -r "$CURRENT_DIR/cmd" "$NEW_PROJECT_DIR/"
cp -r "$CURRENT_DIR/internal" "$NEW_PROJECT_DIR/"
cp -r "$CURRENT_DIR/docs" "$NEW_PROJECT_DIR/"
cp -r "$CURRENT_DIR/tools" "$NEW_PROJECT_DIR/"
cp "$CURRENT_DIR/.env.example" "$NEW_PROJECT_DIR/.env.example"
cp "$CURRENT_DIR/.air.toml" "$NEW_PROJECT_DIR/.air.toml"
cp "$CURRENT_DIR/.gitignore" "$NEW_PROJECT_DIR/.gitignore"
cp "$CURRENT_DIR/README.md" "$NEW_PROJECT_DIR/README.md"

# Create basic .env file
echo "Creating basic .env file..."
cat > "$NEW_PROJECT_DIR/.env" << EOL
# Basic Configuration
PORT=8080
GIN_MODE=debug
LOG_LEVEL=info

# Proxy Configuration
TRUSTED_PROXIES=0.0.0.0/0

# Database Configuration
INIT_DB=false
AUTO_MIGRATE=false
DB_USER=root
DB_PASSWORD=root
DB_HOST=localhost
DB_PORT=3306
DB_NAME=$PROJECT_NAME

# Encryption Configuration
ENCRYPTION_KEY=your_32_byte_encryption_key_here

# Email Configuration
INIT_SMTP=false
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_smtp_username
SMTP_PASSWORD=your_smtp_password
SMTP_FROM_EMAIL=noreply@example.com
EOL

# Create new go.mod file
echo "Creating new go.mod file..."
cd "$NEW_PROJECT_DIR"
go mod init "$PROJECT_NAME"

# Update import paths in all Go files
echo "Updating import paths..."
find . -type f -name "*.go" -exec sed -i '' "s|github.com/cam-boltnote/go-ignite|$PROJECT_NAME|g" {} +

# Update go.mod dependencies
echo "Updating dependencies..."
go mod tidy

# Initialize git repository
echo "Initializing git repository..."
git init

# Create initial commit
git add .
git commit -m "Initial commit"

echo
echo "Project initialization complete!"
echo
echo "Next steps:"
echo "1. cd $NEW_PROJECT_DIR"
echo "2. Update the .env file with your specific configuration values"
echo "3. Set INIT_DB=true in .env if you want to initialize the database"
echo "4. Set INIT_SMTP=true in .env if you want to initialize the email sender"
echo "5. Update TRUSTED_PROXIES in .env with your proxy configuration"
echo "6. Run 'go mod tidy' to download dependencies"
echo "7. Run 'go run cmd/main.go' to start the server"
echo 