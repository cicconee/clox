package cloudstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/db"
	"github.com/google/uuid"
)

// UserValidatorFunc validates that a users directory structure exists
// on the server. It should take a user ID and verify the storage
// mechanism is set up for the user. If it is not, it should set up the
// users storage.
type UserValidatorFunc func(ctx context.Context, userID string) (Dir, error)

// FileService is the business logic for the cloudstore file
// functionality.
//
// FileService should be created using the NewFileService
// function.
type FileService struct {
	path         string
	store        *Store
	io           *IO
	log          *log.Logger
	validateUser UserValidatorFunc
	pathMap      *PathMapper
}

// FileServiceConfig is the FileService configuration.
type FileServiceConfig struct {
	Path         string
	Store        *Store
	IO           *IO
	Log          *log.Logger
	ValidateUser UserValidatorFunc
	PathMap      *PathMapper
}

// NewFileService creates a new FileService.
//
// Path should be the path to the root storage directory on the
// file system. This is the directory that will store all
// directories for Clox users.
//
// Path, Store and ValidateUserDir must be set otherwise it will
// panic.
//
// If IO is not set, it will default to NewIO(&OSFileSystem{}). If
// initializing multiple cloudstore services, it is recommended to
// use the same IO with all services. This will avoid initializing
// multiple IO's.
//
// If PathMap is not set, it will default to NewPathMapper().
//
// If Log is not set, it will default to log.Default().
func NewFileService(c FileServiceConfig) *FileService {
	if c.Path == "" {
		panic("cloudstore.NewFileService: cannot create FileService with empty Path")
	}
	if c.Store == nil {
		panic("cloudstore.NewFileService: cannot create FileService with nil Store")
	}

	if c.ValidateUser == nil {
		panic("cloudstore.NewFileService: cannot create FileService with nil UserValidatorFunc")
	}

	if c.IO == nil {
		c.IO = NewIO(&OSFileSystem{})
	}

	if c.Log == nil {
		c.Log = log.Default()
	}

	if c.PathMap == nil {
		c.PathMap = NewPathMapper()
	}

	return &FileService{
		path:         c.Path,
		store:        c.Store,
		io:           c.IO,
		log:          c.Log,
		validateUser: c.ValidateUser,
		pathMap:      c.PathMap,
	}
}

// FileInfo is the metadata for a file.
type FileInfo struct {
	ID          string
	DirectoryID string
	Name        string
	Path        string
	FSPath      string
	Size        int64
	UploadedAt  time.Time
}

// BatchSave is the result of saving a file when the files are being
// written as a Batch. If an error occured while saving the file it
// will be set in the Err field.
//
// Any BatchSave returned by a service function should always set the
// FileInfo.Name and FileInfo.Size, regardless of an error.
type BatchSave struct {
	FileInfo
	Err error
}

// Msg returns the status of this BatchSave as a user friendly message.
//
// This method is useful to add more context to a response when this
// BatchSave represents a file that failed to be saved.
func (b *BatchSave) Msg() string {
	if b.Err != nil {
		var wrapErr *app.WrappedSafeError
		if errors.As(b.Err, &wrapErr) {
			msg, _ := wrapErr.Safe()
			return msg
		}

		return "Problem writing the file to the server"
	}

	if b.ID != "" {
		return "Success"
	}

	return ""
}

// SaveBatch writes all the files for a user under the specified directory. The
// file permissions are set to 0600. If directoryID is empty, it will default to the
// users root directory. The file names persisted will be the FileName value of each
// multipart.FileHeader.
//
// For each BatchSave that is returned, if an error occured while saving the file, it
// will be set in the Err field. Every BatchSave will always have its Name and Size
// fields set.
//
// SaveBatch validates that a users root directory has been created. If it does not
// exist it will create it.
//
// The file ID and name on the file system will be a randomly generated UUID.
func (s *FileService) SaveBatch(ctx context.Context, userID string, directoryID string, fileHeaders []*multipart.FileHeader) ([]BatchSave, error) {
	root, err := s.validateUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if directoryID == "" {
		directoryID = root.ID
	} else {
		// Ensure the directory exists.
		_, err := s.store.SelectDirectoryByIDUser(ctx, directoryID, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("directory '%s' does not exist: %w", directoryID, err),
					SafeMessage: fmt.Sprintf("Directory '%s' does not exist", directoryID),
					StatusCode:  http.StatusBadRequest,
				})
			}

			return nil, err
		}
	}

	results := []BatchSave{}
	for _, header := range fileHeaders {
		file, err := s.write(ctx, userID, directoryID, header)
		batchSave := BatchSave{FileInfo: file}
		if err != nil {
			batchSave.Err = err
		}

		results = append(results, batchSave)
	}

	return results, nil
}

// SaveBatchPath writes all the files for a user under the specified path. The file
// permissions are set to 0600. The path is cleaned using the filepath.Clean func.
// An empty path will default to the users root directory.
//
// For each BatchSave that is returned, if an error occured while saving the file, it
// will be set in the Err field. Every BatchSave will always have its Name and Size
// fields set.
//
// SaveBatchPath validates that a users root directory has been created. If it does not
// exist it will create it.
//
// The file ID and name on the file system will be a randomly generated UUID.
func (s *FileService) SaveBatchPath(ctx context.Context, userID string, path string, fileHeaders []*multipart.FileHeader) ([]BatchSave, error) {
	root, err := s.validateUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	directoryID, err := s.pathMap.FindDir(ctx, s.store.Query, DirSearch{
		UserID: userID,
		RootID: root.ID,
		Path:   path,
	})
	if err != nil {
		return nil, err
	}

	results := []BatchSave{}
	for _, header := range fileHeaders {
		file, err := s.write(ctx, userID, directoryID, header)
		batchSave := BatchSave{FileInfo: file}
		if err != nil {
			batchSave.Err = err
		}

		results = append(results, batchSave)
	}

	return results, nil
}

// write writes a file to the server and returns a FileInfo. The location of
// the file is defined by the directoryID. Files will be a direct child of the specfied
// directory.
//
// The file write is wrapped in a transaction. If a transaction fails to commit or returns
// an error associated with a inconsistent state, this method will attempt to delete the
// file from the file system. If that fails it will be logged for manual intervention.
//
// The FileInfo returned will always have its Name and Size fields set even if there is
// an error.
func (s *FileService) write(ctx context.Context, userID string, directoryID string, header *multipart.FileHeader) (FileInfo, error) {
	var file FileIO

	err := s.store.Tx(ctx, func(tx *db.Tx) error {
		fileIO, err := s.io.NewFile(ctx, NewQuery(tx), NewFileIO{
			ID:          uuid.NewString(),
			UserID:      userID,
			DirectoryID: directoryID,
			UploadedAt:  time.Now().UTC(),
			Header:      header,
			FSDir:       s.path,
			FSPerm:      0600,
		})
		if err != nil {
			return err
		}

		file = fileIO
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUniqueDirectoryIDName):
			err = app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("file name not available [name: %s, directory_id: %s]: %w", header.Filename, directoryID, err),
				SafeMessage: fmt.Sprintf("File '%s' already exists", header.Filename),
				StatusCode:  http.StatusBadRequest,
			})
		case errors.Is(err, ErrCommitTx), errors.Is(err, ErrCopy):
			go s.removeFS(file.FSPath)
		}

		return FileInfo{Name: header.Filename, Size: header.Size}, err
	}

	return FileInfo{
		ID:          file.ID,
		DirectoryID: file.DirectoryID,
		Name:        file.Name,
		Path:        file.UserPath,
		FSPath:      file.FSPath,
		Size:        file.Size,
		UploadedAt:  file.UploadedAt.UTC(),
	}, nil
}

// removeFS removes a file from the file system. If it fails it will be logged.
func (s *FileService) removeFS(fsPath string) {
	err := s.io.RemoveFS(fsPath)
	if err != nil {
		s.log.Printf("[ERROR] Removing file [path: %s]: %v\n", fsPath, err)
	}
}

// Info gets the information for a users file. It is returned as a FileInfo.
func (s *FileService) Info(ctx context.Context, userID string, fileID string) (FileInfo, error) {
	file, err := s.io.ReadFileInfo(ctx, s.store.Query, ReadFileInfoIO{
		UserID: userID,
		FileID: fileID,
		FSPath: s.path,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FileInfo{}, app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("file '%s' does not exist: %w", fileID, err),
				SafeMessage: "File not found",
				StatusCode:  http.StatusNotFound,
			})
		}

		return FileInfo{}, err
	}

	return FileInfo{
		ID:          file.ID,
		DirectoryID: file.DirectoryID,
		Name:        file.Name,
		Path:        file.UserPath,
		FSPath:      file.FSPath,
		Size:        file.Size,
		UploadedAt:  file.UploadedAt.UTC(),
	}, nil
}
