package external

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GeneralTask/task-manager/backend/constants"
	"github.com/GeneralTask/task-manager/backend/database"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DefaultUserInfoResponse string = `{"data": {"workspaces": [{"gid": "6942069420"}]}}`
)

func TestLoadAsanaTasks(t *testing.T) {
	parentCtx := context.Background()
	db, dbCleanup, err := database.GetDBConnection()
	assert.NoError(t, err)
	defer dbCleanup()
	taskCollection := database.GetTaskCollection(db)

	taskServerSuccess := getMockServer(t, 200, `{"data": [{"gid": "6942069420", "due_on": "2021-04-20", "html_notes": "hmm", "name": "Task!", "permalink_url": "https://example.com/"}]}`)
	userInfoServerSuccess := getMockServer(t, 200, `{"data": {"workspaces": [{"gid": "6942069420"}]}}`)

	t.Run("BadUserInfoStatusCode", func(t *testing.T) {
		userInfoServer := getMockServer(t, 400, "")
		defer userInfoServer.Close()
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{UserInfoURL: &userInfoServer.URL}}}
		userID := primitive.NewObjectID()

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NotEqual(t, nil, result.Error)
		assert.Equal(t, "bad status code: 400", result.Error.Error())
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("BadUserInfoResponse", func(t *testing.T) {
		userInfoServer := getMockServer(t, 200, `oopsie poopsie`)
		defer userInfoServer.Close()
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{UserInfoURL: &userInfoServer.URL}}}
		userID := primitive.NewObjectID()

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NotEqual(t, nil, result.Error)
		assert.Equal(t, "invalid character 'o' looking for beginning of value", result.Error.Error())
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("NoWorkspaceInUserInfo", func(t *testing.T) {
		userInfoServer := getMockServer(t, 200, `{"data": {"workspaces": []}}`)
		defer userInfoServer.Close()
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{UserInfoURL: &userInfoServer.URL}}}
		userID := primitive.NewObjectID()

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NotEqual(t, nil, result.Error)
		assert.Equal(t, "user has not workspaces", result.Error.Error())
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("BadTaskStatusCode", func(t *testing.T) {
		taskServer := getMockServer(t, 409, ``)
		defer taskServer.Close()
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{
			TaskFetchURL: &taskServer.URL,
			UserInfoURL:  &userInfoServerSuccess.URL,
		}}}
		userID := primitive.NewObjectID()

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NotEqual(t, nil, result.Error)
		assert.Equal(t, "bad status code: 409", result.Error.Error())
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("BadTaskResponse", func(t *testing.T) {
		taskServer := getMockServer(t, 200, `to the moon`)
		defer taskServer.Close()
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{
			TaskFetchURL: &taskServer.URL,
			UserInfoURL:  &userInfoServerSuccess.URL,
		}}}
		userID := primitive.NewObjectID()

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NotEqual(t, nil, result.Error)
		assert.Equal(t, "invalid character 'o' in literal true (expecting 'r')", result.Error.Error())
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("Success", func(t *testing.T) {
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{
			TaskFetchURL: &taskServerSuccess.URL,
			UserInfoURL:  &userInfoServerSuccess.URL,
		}}}
		userID := primitive.NewObjectID()

		dueDate, _ := time.Parse("2006-01-02", "2021-04-20")
		expectedTask := database.Task{
			TaskBase: database.TaskBase{
				IDOrdering:      0,
				IDExternal:      "6942069420",
				IDTaskSection:   constants.IDTaskSectionToday,
				Deeplink:        "https://example.com/",
				Title:           "Task!",
				Body:            "hmm",
				SourceID:        TASK_SOURCE_ID_ASANA,
				SourceAccountID: "wrong",
				UserID:          userID,
			},
			DueDate: primitive.NewDateTimeFromTime(dueDate),
		}

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NoError(t, result.Error)
		assert.Equal(t, 1, len(result.Tasks))
		assertTasksEqual(t, &expectedTask, result.Tasks[0])

		var taskFromDB database.Task
		dbCtx, cancel := context.WithTimeout(parentCtx, constants.DatabaseTimeout)
		defer cancel()
		err := taskCollection.FindOne(
			dbCtx,
			bson.M{"user_id": userID},
		).Decode(&taskFromDB)
		assert.NoError(t, err)
		assertTasksEqual(t, &expectedTask, &taskFromDB)
		assert.Equal(t, "sample_account@email.com", taskFromDB.SourceAccountID) // doesn't get updated
	})
	t.Run("SuccessExistingTask", func(t *testing.T) {
		asanaTask := AsanaTaskSource{Asana: AsanaService{ConfigValues: AsanaConfigValues{
			TaskFetchURL: &taskServerSuccess.URL,
			UserInfoURL:  &userInfoServerSuccess.URL,
		}}}
		userID := primitive.NewObjectID()

		dueDate, _ := time.Parse("2006-01-02", "2021-04-20")
		expectedTask := database.Task{
			TaskBase: database.TaskBase{
				IDOrdering:      0,
				IDExternal:      "6942069420",
				IDTaskSection:   constants.IDTaskSectionToday,
				Deeplink:        "https://example.com/",
				Title:           "Task!",
				Body:            "hmm",
				SourceID:        TASK_SOURCE_ID_ASANA,
				SourceAccountID: "sugapapa",
				UserID:          userID,
			},
			DueDate: primitive.NewDateTimeFromTime(dueDate),
		}
		database.GetOrCreateTask(
			db,
			userID,
			"6942069420",
			TASK_SOURCE_ID_ASANA,
			&expectedTask,
		)

		var taskResult = make(chan TaskResult)
		go asanaTask.GetTasks(userID, "sample_account@email.com", taskResult)
		result := <-taskResult
		assert.NoError(t, result.Error)
		assert.Equal(t, 1, len(result.Tasks))
		assertTasksEqual(t, &expectedTask, result.Tasks[0])

		var taskFromDB database.Task
		dbCtx, cancel := context.WithTimeout(parentCtx, constants.DatabaseTimeout)
		defer cancel()
		err := taskCollection.FindOne(
			dbCtx,
			bson.M{"user_id": userID},
		).Decode(&taskFromDB)
		assert.NoError(t, err)
		assertTasksEqual(t, &expectedTask, &taskFromDB)
		assert.Equal(t, "sugapapa", taskFromDB.SourceAccountID) // doesn't get updated
	})
}

func getMockServer(t *testing.T, statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}
