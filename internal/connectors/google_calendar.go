package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarConnector handles Google Calendar API operations
type CalendarConnector struct {
	config *oauth2.Config
}

// NewCalendarConnector creates and initializes a new CalendarConnector
func NewCalendarConnector() (*CalendarConnector, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v", err)
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Read credentials from environment variable
	credentials := os.Getenv("GOOGLE_CALENDAR_CREDENTIALS")
	if credentials == "" {
		return nil, fmt.Errorf("GOOGLE_CALENDAR_CREDENTIALS environment variable not set")
	}

	// Log the configured redirect URIs
	var credsData map[string]interface{}
	if err := json.Unmarshal([]byte(credentials), &credsData); err != nil {
		return nil, fmt.Errorf("failed to parse credentials JSON: %v", err)
	}
	if installed, ok := credsData["installed"].(map[string]interface{}); ok {
		if redirectURIs, ok := installed["redirect_uris"].([]interface{}); ok {
			log.Printf("Configured redirect URIs in credentials: %v", redirectURIs)
		}
	}

	// Parse credentials
	config, err := google.ConfigFromJSON([]byte(credentials),
		calendar.CalendarEventsScope,
		calendar.CalendarReadonlyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret: %v", err)
	}

	// Set the token endpoint URL
	config.Endpoint = google.Endpoint

	log.Printf("OAuth config initialized with redirect URIs: %v", config.RedirectURL)

	return &CalendarConnector{
		config: config,
	}, nil
}

// GetAuthURL generates the OAuth URL for user authorization
func (c *CalendarConnector) GetAuthURL() string {
	// Set the redirect URI to match the frontend's callback URL
	c.config.RedirectURL = "https://app.boltnote.ai/oauth/callback"
	log.Printf("Generating auth URL with redirect URI: %s", c.config.RedirectURL)
	return c.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// Exchange converts an authorization code into a token
func (c *CalendarConnector) Exchange(code string) (*oauth2.Token, error) {
	log.Printf("Starting OAuth exchange with code: %s...", code[:10]) // Show first 10 chars of code

	// Ensure we use the same redirect URI as in GetAuthURL
	c.config.RedirectURL = "https://app.boltnote.ai/oauth/callback"

	token, err := c.config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("OAuth exchange failed: %v", err)
		return nil, fmt.Errorf("failed to exchange auth code: %v", err)
	}

	log.Printf("OAuth exchange successful. Token type: %s, Expiry: %v", token.TokenType, token.Expiry)
	return token, nil
}

// RefreshToken refreshes an expired access token using the refresh token
func (c *CalendarConnector) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := c.config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	return newToken, nil
}

// CreateServiceWithToken creates a new Calendar service with a user's token
func (c *CalendarConnector) CreateServiceWithToken(token *oauth2.Token) (*calendar.Service, error) {
	if token == nil {
		return nil, fmt.Errorf("token cannot be nil")
	}

	client := c.config.Client(context.Background(), token)
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	return srv, nil
}

// CreateEvent creates a new calendar event
func (c *CalendarConnector) CreateEvent(token *oauth2.Token, calendarID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	createdEvent, err := srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %v", err)
	}

	return createdEvent, nil
}

// GetCalendarList retrieves all calendars available to the user
func (c *CalendarConnector) GetCalendarList(token *oauth2.Token) ([]*calendar.CalendarListEntry, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	calendarList, err := srv.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve calendar list: %v", err)
	}

	return calendarList.Items, nil
}

// GetUpcomingEvents retrieves upcoming calendar events
func (c *CalendarConnector) GetUpcomingEvents(token *oauth2.Token, calendarID string, maxResults int64) ([]*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	timeMin := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		MaxResults(maxResults).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %v", err)
	}

	return events.Items, nil
}

// UpdateEvent updates an existing calendar event
func (c *CalendarConnector) UpdateEvent(token *oauth2.Token, calendarID string, eventID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	updatedEvent, err := srv.Events.Update(calendarID, eventID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %v", err)
	}
	return updatedEvent, nil
}

// DeleteEvent deletes a calendar event by ID
func (c *CalendarConnector) DeleteEvent(token *oauth2.Token, calendarID string, eventID string) error {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return err
	}

	err = srv.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete event: %v", err)
	}
	return nil
}

// GetEventByID retrieves a specific event by its ID
func (c *CalendarConnector) GetEventByID(token *oauth2.Token, calendarID string, eventID string) (*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	event, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve event: %v", err)
	}
	return event, nil
}

// GetEventsByTimeRange retrieves events within a specific time range
func (c *CalendarConnector) GetEventsByTimeRange(token *oauth2.Token, calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	events, err := srv.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}
	return events.Items, nil
}

// AddEventReminder adds a reminder to an existing event
func (c *CalendarConnector) AddEventReminder(token *oauth2.Token, calendarID string, eventID string, minutes int64) (*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	event, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, err
	}

	event.Reminders = &calendar.EventReminders{
		UseDefault: false,
		Overrides: []*calendar.EventReminder{
			{
				Method:  "popup",
				Minutes: minutes,
			},
			{
				Method:  "email",
				Minutes: minutes,
			},
		},
	}

	updatedEvent, err := srv.Events.Update(calendarID, eventID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to add reminder: %v", err)
	}
	return updatedEvent, nil
}

// CreateRecurringEvent creates an event that repeats according to a specified frequency
func (c *CalendarConnector) CreateRecurringEvent(token *oauth2.Token, summary, description string, startTime, endTime time.Time, recurrence string) (*calendar.Event, error) {
	srv, err := c.CreateServiceWithToken(token)
	if err != nil {
		return nil, err
	}

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Recurrence: []string{recurrence}, // e.g., "RRULE:FREQ=WEEKLY;COUNT=4"
	}

	recurringEvent, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create recurring event: %v", err)
	}
	return recurringEvent, nil
}

// Helper functions below remain mostly unchanged but are now private to the package

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
