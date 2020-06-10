// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

// Basic interface client of executor module should implement so as to get the
// work done. It has standard id and core execute methods. Also it needs to
// carry the channel with it on which the result of execution will be reported.
// It will also reveal whether it is a blocking task or not.
type Task interface {
	GetId() int

	Execute() Response

	SetRespChan(rc chan Response)

	GetRespChan() chan Response

	IsBlocking() bool
}
