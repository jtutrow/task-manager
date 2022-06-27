package external

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/GeneralTask/task-manager/backend/constants"
	"github.com/GeneralTask/task-manager/backend/database"
	"github.com/GeneralTask/task-manager/backend/utils"

	"github.com/stretchr/testify/assert"
)

func TestLoadSlackTasks(t *testing.T) {
	userID := primitive.NewObjectID()
	task := createTestSlackTask(userID)
	taskWrongSource := createTestSlackTask(userID)
	taskWrongSource.SourceID = TASK_SOURCE_ID_GCAL
	taskCompleted := createTestSlackTask(userID)
	taskCompleted.IsCompleted = true
	insertTestTasks(
		t,
		userID,
		[]*database.Item{
			task,
			taskWrongSource,
			taskCompleted,
		},
	)

	t.Run("Success", func(t *testing.T) {
		var tasks = make(chan TaskResult)
		go SlackSavedTaskSource{}.GetTasks(userID, GeneralTaskDefaultAccountID, tasks)
		result := <-tasks
		assert.NoError(t, result.Error)
		assert.Equal(t, 1, len(result.Tasks))
		// check IDExternal because ID is set upon db insertion
		assert.Equal(t, task.IDExternal, result.Tasks[0].IDExternal)
	})
	t.Run("WrongUserID", func(t *testing.T) {
		var tasks = make(chan TaskResult)
		go SlackSavedTaskSource{}.GetTasks(primitive.NewObjectID(), GeneralTaskDefaultAccountID, tasks)
		result := <-tasks
		assert.NoError(t, result.Error)
		assert.Equal(t, 0, len(result.Tasks))
	})
	t.Run("WrongSourceAccountID", func(t *testing.T) {
		var tasks = make(chan TaskResult)
		go SlackSavedTaskSource{}.GetTasks(userID, "other_account_id", tasks)
		result := <-tasks
		assert.NoError(t, result.Error)
		assert.Equal(t, 0, len(result.Tasks))
	})
}

func TestCreateSlackTask(t *testing.T) {
	db, dbCleanup, err := database.GetDBConnection()
	assert.NoError(t, err)
	defer dbCleanup()

	testTask := database.Item{
		TaskBase: database.TaskBase{
			IDTaskSection:   constants.IDTaskSectionDefault,
			SourceID:        TASK_SOURCE_ID_SLACK_SAVED,
			Title:           "send dogecoin to the moon",
			Body:            "",
			SourceAccountID: GeneralTaskDefaultAccountID,
		},
		TaskType: database.TaskType{
			IsTask: true,
		},
		SlackMessageParams: database.SlackMessageParams{
			Channel:  "test channel",
			SenderID: "test sender ID",
			Team:     "test team",
			TimeSent: "test ts",
		},
	}
	t.Run("SuccessSlackCreation", func(t *testing.T) {
		userID := primitive.NewObjectID()
		testTask.TaskBase.UserID = userID
		_, err := SlackSavedTaskSource{}.CreateNewTask(userID, GeneralTaskDefaultAccountID, TaskCreationObject{
			Title: "send dogecoin to the moon",
			SlackMessageParams: database.SlackMessageParams{
				Channel:  "test channel",
				SenderID: "test sender ID",
				Team:     "test team",
				TimeSent: "test ts",
			},
		})
		assert.NoError(t, err)
		tasks, err := database.GetActiveTasks(db, userID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(*tasks))
		task := (*tasks)[0]
		assert.True(t, task.IsTask)
		utils.AssertTasksEqual(t, &task, &testTask)
	})
	t.Run("SuccessSlackCustomSpecifySection", func(t *testing.T) {
		userID := primitive.NewObjectID()
		testTask.TaskBase.UserID = userID
		taskSectionID := primitive.NewObjectID()
		testTask.IDTaskSection = taskSectionID
		_, err := SlackSavedTaskSource{}.CreateNewTask(userID, GeneralTaskDefaultAccountID, TaskCreationObject{
			Title: "send dogecoin to the moon",
			SlackMessageParams: database.SlackMessageParams{
				Channel:  "test channel",
				SenderID: "test sender ID",
				Team:     "test team",
				TimeSent: "test ts",
			},
			IDTaskSection: taskSectionID,
		})
		assert.NoError(t, err)
		tasks, err := database.GetActiveTasks(db, userID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(*tasks))
		task := (*tasks)[0]
		assert.True(t, task.IsTask)
		utils.AssertTasksEqual(t, &task, &testTask)
	})
}

func createTestSlackTask(userID primitive.ObjectID) *database.Item {
	return &database.Item{
		TaskBase: database.TaskBase{
			IDOrdering:      2,
			IDExternal:      primitive.NewObjectID().Hex(),
			IDTaskSection:   constants.IDTaskSectionDefault,
			Title:           "Sample Taskeroni",
			SourceID:        TASK_SOURCE_ID_SLACK_SAVED,
			UserID:          userID,
			SourceAccountID: GeneralTaskDefaultAccountID,
		},
	}
}
