package helpers

import "os"

func EnsureDir(dir string) error {
	// Создаем директорию, если она не существует
	err := os.MkdirAll(dir, 0755) // 0755 — права доступа
	if err != nil {
		return err
	}
	return nil
}
