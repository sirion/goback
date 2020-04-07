package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

///
/// GoBack Tests
///

func TestEmpty(t *testing.T) {
	backup := &Backup{}

	args := createTestEnv(t)

	exit := backup.loadConfiguration(args)
	if exit != nil {
		t.Fatalf("Exited backup.loadConfiguration with code %d: %s", exit.Code, exit.Message)
	}

	exit = backup.hash()
	if exit != nil {
		t.Fatalf("Exited backup.hash with code %d: %s", exit.Code, exit.Message)
	}

	if !backup.Initial {
		t.Error("Backup not set to initial")
	}

	if backup.FromHashes == nil {
		t.Error("FromHashes not initialized")
	}

	if len(backup.FromHashes) != 0 {
		t.Error("FromHashes not empty")
	}

	if backup.Ref != "" {
		t.Error("Ref not empty")
	}

	if backup.RefHashes == nil {
		t.Error("RefHashes not initialized")
	}

	if len(backup.RefHashes) != 0 {
		t.Error("RefHashes not empty")
	}

	cleanupTestEnv(args)

}

func TestMetadataInitial(t *testing.T) {
	backup := &Backup{}

	args := createTestEnv(t)

	createTestFiles(t, []string{
		filepath.Join(args.Source, "test01"),
		filepath.Join(args.Source, "dir/test02"),
	})

	exit := backup.loadConfiguration(args)
	if exit != nil {
		t.Fatalf("Exited backup.loadConfiguration with code %d: %s", exit.Code, exit.Message)
	}

	exit = backup.hash()
	if exit != nil {
		t.Fatalf("Exited backup.hash with code %d: %s", exit.Code, exit.Message)
	}

	if !backup.Initial {
		t.Error("Backup not set to initial")
	}

	if len(backup.FromHashes) != 2 {
		t.Error("FromHashes not correctly filled")
	}

	if backup.Ref != "" {
		t.Error("Ref not empty")
	}

	if len(backup.RefHashes) != 0 {
		t.Error("RefHashes not nil")
	}

	cleanupTestEnv(args)

}

func TestBackupEmpty(t *testing.T) {
	backup := &Backup{}

	args := createTestEnv(t)

	exit := backup.loadConfiguration(args)
	if exit != nil {
		t.Fatalf("Exited backup.loadConfiguration with code %d: %s", exit.Code, exit.Message)
	}

	exit = backup.hash()
	if exit != nil {
		t.Fatalf("Exited backup.hash with code %d: %s", exit.Code, exit.Message)
	}

	exit = backup.create()
	if exit != nil {
		t.Fatalf("Exited backup.create with code %d: %s", exit.Code, exit.Message)
	}

	cleanupTestEnv(args)
}

func TestBackupSequence(t *testing.T) {

	args := createTestEnv(t)

	fmt.Fprintln(os.Stderr, "--== Check 1 ==--")
	createTestFiles(t, []string{
		filepath.Join(args.Source, "test01"),
	})

	backupAndAssert(t, BackupAssertion{
		Prefix:           "Check 1",
		IsInitial:        true,
		NumBackups:       1,
		NumBackupFolders: 1,
		FilesBackup:      []string{"test01"},
		FilesRefBefore:   []string{},
		FilesRefAfter:    []string{},
	}, args)

	time.Sleep(10 * time.Millisecond)

	fmt.Fprintln(os.Stderr, "--== Check 2 ==--")
	createTestFiles(t, []string{
		filepath.Join(args.Source, "test02"),
	})

	backupAndAssert(t, BackupAssertion{
		Prefix:           "Check 2",
		IsInitial:        false,
		NumBackups:       2,
		NumBackupFolders: 1,
		FilesBackup:      []string{"test01", "test02"},
		FilesRefBefore:   []string{"test01"},
		FilesRefAfter:    []string{},
	}, args)

	time.Sleep(10 * time.Millisecond)

	fmt.Fprintln(os.Stderr, "--== Check 3 ==--")
	createTestFiles(t, []string{
		filepath.Join(args.Source, "test03/a"),
		filepath.Join(args.Source, "test03/b"),
		filepath.Join(args.Source, "test03/c"),
	})

	backupAndAssert(t, BackupAssertion{
		Prefix:           "Check 3",
		IsInitial:        false,
		NumBackups:       3,
		NumBackupFolders: 1,
		FilesBackup:      []string{"test01", "test02", "test03/a", "test03/b", "test03/c"},
		FilesRefBefore:   []string{"test01", "test02"},
		FilesRefAfter:    []string{},
	}, args)

	time.Sleep(10 * time.Millisecond)

	fmt.Fprintln(os.Stderr, "--== Check 4 ==--")
	err := os.Remove(filepath.Join(args.Source, "test03/b"))
	if err != nil {
		t.Fatalf("Error removing test file %s: %s", filepath.Join(args.Source, "test03/b"), err.Error())
	}

	backupAndAssert(t, BackupAssertion{
		Prefix:           "Check 4",
		IsInitial:        false,
		NumBackups:       4,
		NumBackupFolders: 2,
		FilesBackup:      []string{"test01", "test02", "test03/a", "test03/c"},
		FilesRefBefore:   []string{"test01", "test02", "test03/a", "test03/b", "test03/c"},
		FilesRefAfter:    []string{"test03/b"},
	}, args)

	cleanupTestEnv(args)
}

///
/// Helper Functions
///

type BackupAssertion struct {
	Prefix           string
	IsInitial        bool
	FilesBackup      []string
	FilesRefBefore   []string
	FilesRefAfter    []string
	NumBackupFolders int
	NumBackups       int
}

func backupAndAssert(t *testing.T, assert BackupAssertion, args *Arguments) {
	backup := &Backup{}

	exit := backup.loadConfiguration(args)
	if exit != nil {
		t.Fatalf("[%s] Exited backup.loadConfiguration with code %d: %s", assert.Prefix, exit.Code, exit.Message)
	}

	if assert.IsInitial != backup.Initial {
		t.Errorf("[%s] Backup initial should be %t but is %t", assert.Prefix, assert.IsInitial, backup.Initial)
	}

	if backup.FromHashes != nil {
		t.Errorf("[%s] FromHashes were filled before backup", assert.Prefix)
	}
	if backup.RefHashes != nil {
		t.Errorf("[%s] FromHashes were filled before backup", assert.Prefix)
	}

	exit = backup.hash()
	if exit != nil {
		t.Errorf("[%s] Exited backup.hash with code %d: %s", assert.Prefix, exit.Code, exit.Message)
	}

	if len(assert.FilesBackup) != len(backup.FromHashes) {
		t.Errorf("[%s] Number of hashes in backup not correct. Is: %d, should be %d", assert.Prefix, len(backup.FromHashes), len(assert.FilesBackup))
	} else {
		for _, file := range assert.FilesBackup {
			_, ok := backup.FromHashes[file]
			if !ok {
				t.Errorf("[%s] Hash not found in backup: %s", assert.Prefix, file)
			}
		}
	}

	if len(assert.FilesRefBefore) != len(backup.RefHashes) {
		t.Errorf("[%s] Number of hashes in reference before backup not correct. Is: %d, should be %d", assert.Prefix, len(backup.RefHashes), len(assert.FilesRefBefore))
	} else {
		for _, file := range assert.FilesRefBefore {
			_, ok := backup.RefHashes[file]
			if !ok {
				t.Errorf("[%s] Hash not found in reference before backup: %s", assert.Prefix, file)
			}
		}
	}

	// fmt.Printf("Ref folder: %s\n", backup.Ref)
	exit = backup.create()
	if exit != nil {
		t.Fatalf("[%s] Exited backup.create with code %d: %s", assert.Prefix, exit.Code, exit.Message)
	}

	backups := listBackups(backup.Configuration.targetDirectory)
	if len(backups) != assert.NumBackups {
		backupFiles := fmt.Sprintf("Backups: \n - %s\n", strings.Join(backups, "\n - "))
		t.Fatalf("[%s] Wrong number of backups: Is: %d, should be %d: %s", assert.Prefix, len(backups), assert.NumBackups, backupFiles)
	}

	backupDirs, err := listDirs(backup.Configuration.targetDirectory)
	if err != nil {
		t.Fatalf("[%s] Error listing backup folders: %s", assert.Prefix, err.Error())
	}

	if len(backupDirs) != assert.NumBackupFolders {
		backupFolders := fmt.Sprintf("Folders in backup: \n - %s\n", strings.Join(backupDirs, "\n - "))
		t.Fatalf("[%s] Wrong number of backup folders: Is: %d, should be %d: %s", assert.Prefix, len(backupDirs), assert.NumBackupFolders, backupFolders)
	}

	backupFiles := listFiles(backup.To)
	if len(assert.FilesBackup) != len(backupFiles) {
		fileMessage := fmt.Sprintf("Files in backup: \n - %s\n", strings.Join(backupFiles, "\n - "))
		t.Errorf("[%s] Number of files in backup not correct. Is: %d, should be %d\n%s", assert.Prefix, len(backupFiles), len(assert.FilesBackup), fileMessage)
	} else {
		for _, searchFile := range assert.FilesBackup {
			found := false
			for _, file := range backupFiles {
				if file == searchFile {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%s] File not found in backup: %s", assert.Prefix, searchFile)
			}
		}
	}

	var referenceFiles []string
	if backup.Ref == "" {
		referenceFiles = []string{}
	} else {
		referenceFiles = listFiles(backup.Ref)
	}
	fmt.Printf("Ref files in %s: %s\n", backup.Ref, strings.Join(referenceFiles, ", "))
	if len(assert.FilesRefAfter) != len(referenceFiles) {
		t.Errorf("[%s] Number of files in reference after backup not correct. Is: %d, should be %d", assert.Prefix, len(referenceFiles), len(assert.FilesRefAfter))
	} else {
		for _, searchFile := range assert.FilesRefAfter {
			found := false
			for _, file := range referenceFiles {
				if file == searchFile {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%s] File not found in reference after backup: %s", assert.Prefix, searchFile)
			}
		}
	}

}

func createTestEnv(t *testing.T) *Arguments {
	pathSourceDir, pathBackupDir, err := createTestDirs()
	if err != nil {
		t.Fatal(err.Error())
	}

	args := &Arguments{
		// ChangeDetection: ChangeDetectionModificationAndSize,
		OutputLevel: OutputLevelDebug,
		Source:      pathSourceDir,
		Target:      pathBackupDir,
		Type:        BackupTypeTest,
	}

	return args
}

func cleanupTestEnv(args *Arguments) {
	_ = os.RemoveAll(filepath.Dir(args.Source))
}

func listBackups(dir string) []string {
	suffix := ".goback"

	dirs := make([]string, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return dirs
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		name := info.Name()
		if name != ConfigurationFile && strings.HasSuffix(name, suffix) {
			dirs = append(dirs, info.Name())
		}
	}

	return dirs

}

func listFiles(dir string) []string {
	fileNames := make([]string, 0, 1)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			fileNames = append(fileNames, path[len(dir)+1:])
		}
		return nil
	})

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error walking directory %s: %s", dir, err.Error())
	}

	return fileNames
}

func listDirs(dir string) ([]string, error) {
	dirs := make([]string, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			dirs = append(dirs, info.Name())
		}
	}

	return dirs, nil
}

func createTestDirs() (string, string, error) {
	pathMainDir, err := ioutil.TempDir("", "goback*")
	if err != nil {
		return "", "", fmt.Errorf("Cannot create temp main dir: %s", err.Error())
	}

	pathSourceDir := filepath.Join(pathMainDir, "source")
	pathBackupDir := filepath.Join(pathMainDir, "backup")

	err = os.Mkdir(pathSourceDir, os.ModePerm)
	if err != nil {
		return "", "", fmt.Errorf("Cannot create temp source dir: %s", err.Error())
	}

	err = os.Mkdir(pathBackupDir, os.ModePerm)
	if err != nil {
		return "", "", fmt.Errorf("Cannot create temp backup dir: %s", err.Error())
	}

	return pathSourceDir, pathBackupDir, nil
}

const MaxFileSize = 1024

func createTestFile(pathFile string) error {
	dir, _ := filepath.Split(pathFile)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic("Could not create directory for test file")
	}

	size := rand.Intn(MaxFileSize)

	buffer := make([]byte, size)
	rand.Read(buffer)

	return ioutil.WriteFile(pathFile, buffer, os.ModePerm)
}

func createTestFiles(t *testing.T, paths []string) {
	for _, path := range paths {
		err := createTestFile(path)
		if err != nil {
			t.Fatalf("Error creating file %s: %s", path, err.Error())
		}
	}
}
