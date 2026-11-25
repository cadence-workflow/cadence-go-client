package internal

// CreateSemaphoreRequest is the request to SemaphoreClient.Create.
type CreateSemaphoreRequest struct {
	// SemaphoreName is unique within the domain. Required.
	SemaphoreName string
	// Capacity is the number of tokens, which is how many holders the semaphore allows at once. Required.
	Capacity int32
	// BucketCapacity is the number of tokens in each bucket. Optional; the server picks a default when
	// it is 0.
	BucketCapacity int32
}

// Semaphore is a semaphore returned by SemaphoreClient.
type Semaphore struct {
	SemaphoreName  string
	Capacity       int32
	BucketCapacity int32
}
