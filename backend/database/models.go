package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// https://www.mongodb.com/blog/post/quick-start-golang--mongodb--modeling-documents-with-go-data-structures

// APISource is a distinct API with its own token
type APISource string

const (
	// Google APISource
	Google APISource = "google"
)

// User model
// todo: consider putting api tokens into user document
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	GoogleID string             `bson:"google_id"`
	Email    string             `bson:"email"`
	Name     string             `bson:"name"`
}

// InternalAPIToken model
type InternalAPIToken struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Token  string             `bson:"token"`
	UserID primitive.ObjectID `bson:"user_id"`
}

// ExternalAPIToken model
type ExternalAPIToken struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Source       string             `bson:"source"`
	Token        string             `bson:"token"`
	UserID       primitive.ObjectID `bson:"user_id"`
	AccountID    string             `bson:"account_id"`
	DisplayID    string             `bson:"display_id"`
	IsUnlinkable bool               `bson:"is_unlinkable"`
}

type AtlassianSiteConfiguration struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	UserID  primitive.ObjectID `bson:"user_id"`
	CloudID string             `bson:"cloud_id"`
	SiteURL string             `bson:"site_url"`
}

type JIRAPriority struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id"`
	JIRAID          string             `bson:"jira_id"`
	IntegerPriority int                `bson:"integer_priority"`
}

type StateToken struct {
	Token  primitive.ObjectID `bson:"_id,omitempty"`
	UserID primitive.ObjectID `bson:"user_id"`
}

// Task json & mongo model
type TaskBase struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID           primitive.ObjectID `json:"-" bson:"user_id"`
	IDExternal       string             `json:"-" bson:"id_external"`
	IDOrdering       int                `json:"id_ordering" bson:"id_ordering"`
	IDTaskSection    primitive.ObjectID `json:"-" bson:"id_task_section"`
	IsCompleted      bool               `json:"-" bson:"is_completed"`
	Sender           string             `json:"sender" bson:"sender"`
	Source           TaskSource         `json:"source" bson:"source"`
	SourceAccountID  string             `json:"-" bson:"source_account_id"`
	Deeplink         string             `json:"deeplink" bson:"deeplink"`
	Title            string             `json:"title" bson:"title"`
	Body             string             `json:"body" bson:"body"`
	HasBeenReordered bool               `json:"-" bson:"has_been_reordered"`
	//time in nanoseconds
	TimeAllocation int64 `json:"time_allocated" bson:"time_allocated"`
}

type CalendarEvent struct {
	TaskBase      `bson:",inline"`
	DatetimeEnd   primitive.DateTime `bson:"datetime_end"`
	DatetimeStart primitive.DateTime `bson:"datetime_start"`
}

type CalendarEventChangeableFields struct {
	Title         string             `json:"title" bson:"title,omitempty"`
	DatetimeEnd   primitive.DateTime `bson:"datetime_end,omitempty"`
	DatetimeStart primitive.DateTime `bson:"datetime_start,omitempty"`
}

type Email struct {
	TaskBase     `bson:",inline"`
	ThreadID     string             `bson:"thread_id"`
	SenderDomain string             `bson:"sender_domain"`
	TimeSent     primitive.DateTime `bson:"time_sent"`
}

type Task struct {
	TaskBase           `bson:",inline"`
	DueDate            primitive.DateTime `bson:"due_date"`
	PriorityID         string             `bson:"priority_id"`
	PriorityNormalized float64            `bson:"priority_normalized"`
	TaskNumber         int                `bson:"task_number"`
}

type TaskChangeableFields struct {
	Title              string             `json:"title" bson:"title,omitempty"`
	DueDate            primitive.DateTime `bson:"due_date,omitempty"`
	PriorityID         string             `bson:"priority_id,omitempty"`
	PriorityNormalized float64            `bson:"priority_normalized,omitempty"`
}

type TaskSource struct {
	Name          string `json:"name" bson:"name"`
	Logo          string `json:"logo" bson:"logo"`
	IsCompletable bool   `json:"is_completable" bson:"is_completable"`
	IsReplyable   bool   `json:"is_replyable" bson:"is_replyable"`
}

var TaskSourceGoogleCalendar = TaskSource{
	"Google Calendar",
	"/images/gcal.svg",
	false,
	false,
}

var TaskSourceGmail = TaskSource{
	"Gmail",
	"/images/gmail.svg",
	true,
	true,
}
var TaskSourceJIRA = TaskSource{
	"Jira",
	"/images/jira.svg",
	true,
	false,
}

var TaskSourceSlack = TaskSource{
	"Slack",
	"/images/slack.svg",
	true,
	true,
}

var TaskSourceNameToSource = map[string]TaskSource{
	TaskSourceGoogleCalendar.Name: TaskSourceGoogleCalendar,
	TaskSourceGmail.Name:          TaskSourceGmail,
	TaskSourceJIRA.Name:           TaskSourceJIRA,
	TaskSourceSlack.Name:          TaskSourceSlack,
	// Add "google" so this map can be used for external API token source also
	"google": TaskSourceGmail,
}

type UserSetting struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"`
	FieldKey   string             `bson:"field_key"`
	FieldValue string             `bson:"field_value"`
}

type WaitlistEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	HasAccess bool               `bson:"has_access"`
	CreatedAt primitive.DateTime `bson:"created_at"`
}
