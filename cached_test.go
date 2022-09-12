package logging

// func TestNewWithCancel(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	mockLogger := &logger.MockLogger{}

// 	mockLogger.On("Close").Return(nil)

// 	cached := NewCached[any](
// 		mockLogger,
// 		WithBufferSize(100),
// 		WithWorkerPool(5),
// 		WithFlushRate(100),
// 		WithRetryCount(0),
// 	)

// 	assert.NotNil(cached)
// 	assert.NoError(cached.Close())
// 	mockLogger.AssertExpectations(t)
// }
