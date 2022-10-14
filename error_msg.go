package logger

const (
	notEnoughBytesWritten    = `{"msg":"failed to write all data to the writer","actualLen":%d,"expectedLen":%d}`
	failedToWriteToTheFile   = `{"msg":"failed to write to the file %s","error":"%v"}`
	failedToCloseTheFile     = `{"msg":"failed to close the file %s","error":"%v"}`
	failedToOpenFile         = `{"msg":"failed to open the file %s","error":"%v"}`
	failedToSerializeTheData = `{"msg":"failed to serialize the data","error":"%v"}`
)
