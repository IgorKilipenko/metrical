package testutils

// TestFilePaths содержит пути к тестовым файлам
const (
	TestMetricsFile = "/tmp/test-metrics.json"
	TestSyncFile    = "/tmp/test-sync-metrics.json"
	TestEnvFile     = "/tmp/test-env-metrics.json"
	TestFlagFile    = "/tmp/test-flag-metrics.json"
)

// GetTestFilePath возвращает путь к тестовому файлу с указанным именем
func GetTestFilePath(name string) string {
	return "/tmp/test-" + name + "-metrics.json"
}
