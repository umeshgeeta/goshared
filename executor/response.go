// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

const TaskStatusNotSubmitted = 0
const TaskStatusFailedToSubmit = 1
const TaskStatusSubmitted = 100
const TaskStatusCompletedSuccessfully = 200
const TaskStatusCompletedFailed = 500

type Response struct {
	// Id of the task to which this response corresponds to
	TaskId int

	// Task status - whether it succeeded or failed or some other state
	Status int

	// Output of the successfully executed task.
	// For now we assume it will be a JSON string
	Result string

	Errors []error
}

func NewResponse(tid int) *Response {
	r := new(Response)
	r.TaskId = tid
	return r
}

func FailedToSubmitResponse(tid int) *Response {
	r := NewResponse(tid)
	r.Status = TaskStatusFailedToSubmit
	return r
}
