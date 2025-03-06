# Go Activity Tracking API

A RESTful API service built with Go (Gin) for tracking user activities with features for user management, settings management, and more.

## Features

- User Management (registration, authentication, profile management)
- User Settings Management (notifications, privacy, general preferences)
- Database Integration with MySQL
- JWT Authentication
- CORS Support
- Configurable Auto-migrations
- Email Notifications
- Custom User Settings Support

## Prerequisites

- Go 1.16 or higher
- MySQL 5.7 or higher
- SMTP server for email notifications (optional)

## Connectors

The application uses connectors as pure integration layers for external services. Connectors provide basic operations without business logic and are meant to be used by services rather than directly.

### Architecture

```
internal/
├── connectors/     # Integration layer
│   ├── mysql.go    # Database operations
│   ├── email.go    # Raw email sending
│   ├── openai.go   # AI API integration
│   ├── calendar.go # Calendar operations
│   └── weaviate.go # Vector DB operations
├── services/       # Business logic layer
│   ├── user.go     # Uses mysql, email connectors
│   ├── auth.go     # Uses mysql connector
│   ├── ai.go       # Uses openai, weaviate connectors
│   └── calendar.go # Uses calendar connector
└── routes/         # API layer
```

## Getting Started

Follow these steps to get the project up and running:

1. **Clone the Repository**
   ```bash
   git clone <repository-url>
   cd go-ignite
   ```

2. **Set Up Environment Variables**
   ```bash
   # Copy the example environment file
   cp .env.example .env
   ```

   At minimum, you need to set these basic variables in your `.env` file:
   ```env
   # Server Configuration
   PORT=8080
   GIN_MODE=debug
   
   # Database Configuration
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_HOST=localhost
   DB_PORT=3306
   DB_NAME=your_database_name
   
   # Logging Configuration
   LOG_LEVEL=info
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   go install github.com/air-verse/air@latest
   ```

4. **Verify Project Structure**
   Ensure you have these key directories:
   ```
   .
   ├── cmd/
   │   └── main.go
   ├── internal/
   │   ├── routes/
   │   ├── services/
   │   └── middleware/
   ├── go.mod
   ├── go.sum
   └── .env
   ```

5. **Run the Server**
   ```bash
   # Option 1: Regular run
   go run cmd/main.go

   # Option 2: Hot reload with Air (recommended for development)
   air
   ```

6. **Test the API**
   ```bash
   # Using curl
   curl http://localhost:8080/api/v1/test
   
   # Expected response:
   # {"message":"the work is mysterious and important"}
   ```

   Or use your preferred API testing tool (Postman, Insomnia, etc.) to make a GET request to:
   ```
   http://localhost:8080/api/v1/test
   ```

### Troubleshooting

- If you get a port conflict error, modify the `PORT` in your `.env` file
- If database connection fails, verify your database credentials in the `.env` file
- Ensure all Go dependencies are properly downloaded
- Check that your `GOPATH` is correctly set
- Verify that the server started without any errors in the console

### Database Connector (MySQL)
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_database_name
```

Usage example:
```go
import "github.com/cam-boltnote/go-ignite/internal/connectors"

db, err := connectors.NewDatabase()
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Use the database
result := db.GetDB().Create(&someModel)
```

### Email Connector (SMTP)
```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_smtp_username
SMTP_PASSWORD=your_smtp_password
SMTP_FROM_EMAIL=noreply@example.com
```

Usage example:
```go
emailClient, err := connectors.NewEmailSender()
if err != nil {
    log.Fatal(err)
}

err = emailClient.SendEmail("recipient@example.com", "Subject", "Email body")
```

### OpenAI Connector
```env
OPENAI_API_KEY=your-api-key
OPENAI_DEFAULT_MODEL=gpt-3.5-turbo  # optional
OPENAI_DEFAULT_TEMPERATURE=0.7      # optional
```

Usage example:
```go
client, err := connectors.NewOpenAIClient()
if err != nil {
    log.Fatal(err)
}

messages := []connectors.ChatMessage{
    {Role: "user", Content: "Hello!"},
}
response, err := client.CreateUnstructuredChatCompletion(messages, "", nil)
```

### Google Calendar Connector
```env
GOOGLE_CALENDAR_CREDENTIALS={"web":{"client_id":"...","client_secret":"...",...}}
```

Usage example:
```go
calendar, err := connectors.NewCalendarConnector()
if err != nil {
    log.Fatal(err)
}

authURL := calendar.GetAuthURL()
// Handle OAuth flow...
```

### Weaviate Connector (Vector Database)
```env
WEAVIATE_HOST=your-cluster.weaviate.network
WEAVIATE_API_KEY=your-api-key
OPENAI_API_KEY=your-openai-key  # Required for embeddings
```

Usage example:
```go
client, err := connectors.NewWeaviateClient()
if err != nil {
    log.Fatal(err)
}

// Create embeddings and store in Weaviate
vector, err := client.CreateEmbedding("Some text")
if err != nil {
    log.Fatal(err)
}

err = client.AddObject(ctx, "ClassName", properties, vector)
```

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure the following variables:

```env
# Server Configuration
PORT=8080                   # API server port
GIN_MODE=debug             # gin mode (debug/release)
ENCRYPTION_KEY=            # 32-byte encryption key for JWT
AUTO_MIGRATE=false         # Enable/disable automatic database migrations

# Database Configuration
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=your_database_name

# Email Configuration (Optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_smtp_username
SMTP_PASSWORD=your_smtp_password
SMTP_FROM_EMAIL=noreply@example.com

# System Configuration
PASSWORD_MIN_LENGTH=8      # Minimum password length
PASSWORD_MAX_LENGTH=72     # Maximum password length
ADMIN_NOTIFICATION_EMAIL=  # Email for admin notifications

# Logging Configuration
LOG_LEVEL=info            # Logging level (debug, info, warn, error, fatal)
```

## Logging

The application uses structured logging with level-based file outputs. Logs are stored in the `/logs` directory with separate files for each log level:

- `debug.log` - Detailed debugging information
- `info.log` - General operational information
- `warn.log` - Warning messages for potentially harmful situations
- `error.log` - Error messages for serious problems
- `fatal.log` - Critical errors that require immediate attention

Example service usage:
```go
type ExampleService struct {
    logger *utils.Logger
}

func NewExampleService() *ExampleService {
    return &ExampleService{
        logger: utils.GetLogger().WithService("example_service"),
    }
}

func (s *ExampleService) DoSomething() error {
    s.logger.Info("Starting operation", map[string]interface{}{
        "timestamp": time.Now(),
    })
    
    if err := someOperation(); err != nil {
        s.logger.Error("Operation failed", err, map[string]interface{}{
            "details": "Additional context",
        })
        return err
    }
    
    return nil
}
```

## API Endpoints

### Public Routes

- `POST /api/v1/user` - Create new user
- `POST /api/v1/user/login` - User login

### Test Routes
- `GET /api/v1/test` - Get test message (returns a simple test message)
- `OPTIONS /api/v1/test` - CORS preflight for test endpoint

### Protected Routes (Requires Authentication)

#### User Management
- `GET /api/v1/user/:id` - Get user details
- `PUT /api/v1/user/:id` - Update user
- `DELETE /api/v1/user/:id` - Delete user
- `GET /api/v1/user/email/:email` - Get user by email
- `PUT /api/v1/user/:id/password` - Update password
- `PUT /api/v1/user/:id/activate` - Activate user
- `PUT /api/v1/user/:id/deactivate` - Deactivate user

#### Settings Management
- `GET /api/v1/settings/:userId` - Get user settings
- `PUT /api/v1/settings/:userId` - Update user settings
- `PUT /api/v1/settings/:userId/notifications` - Update notification settings
- `PUT /api/v1/settings/:userId/privacy` - Update privacy settings
- `PUT /api/v1/settings/:userId/general` - Update general settings
- `PUT /api/v1/settings/:userId/custom` - Update custom settings
- `GET /api/v1/settings/:userId/custom/:key` - Get specific custom setting

### Health Check
- `GET /api/v1/health` - API health check

## Database Models

### User Model
- ID (uint)
- Email (string, unique)
- Password (string)
- FirstName (string)
- LastName (string)
- Role (string)
- IsActive (bool)
- CreatedAt (time.Time)
- UpdatedAt (time.Time)
- DeletedAt (gorm.DeletedAt)

### Settings Model
- UserID (uint, unique)
- Timezone (string, default: "UTC")
- Language (string, default: "en")
- Theme (string, default: "light")
- EmailNotificationsEnabled (bool, default: true)
- PushNotificationsEnabled (bool, default: true)
- NotificationFrequency (string, default: "daily")
- ProfileVisibility (string, default: "private")
- DataSharing (bool, default: false)
- CustomSettings (JSON)

## Development

1. Clone the repository
2. Copy `.env.example` to `.env` and configure the variables
3. Install dependencies:
   ```bash
   go mod download
   go install github.com/air-verse/air@latest
   ```
4. Run the server (choose one):
   ```bash
   # Option 1: Regular run
   go run cmd/main.go

   # Option 2: Hot reload with Air (recommended for development)
   air
   ```

The project includes Air configuration for hot reloading during development. When you run `go mod download`, Air will be automatically added to your tools. You can then use `air` command to start the server with hot reload.

## Security Notes

- Passwords are currently stored in plaintext. In production, implement proper password hashing.
- JWT tokens should be properly secured with appropriate expiration times.
- CORS settings should be configured according to your production environment.
- API rate limiting should be implemented for production use.

## Error Handling

The API returns standardized error responses in JSON format:
```json
{
    "error": "Error message description"
}
```

## Database Connection

The application supports lazy loading of the database connection. If the database is not available, the application will still start, but database-dependent features will be disabled. The test endpoint will remain accessible regardless of database availability.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
