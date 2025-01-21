package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	HistoryRecordCount = 30_000

	TargetURL = ""

	InitialID = 300_000

	FolderName = "New folder"
)

type HistoryRecord struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Bookmarks struct {
	Checksum     string `json:"checksum"`
	Roots        Roots  `json:"roots"`
	SyncMetadata string `json:"sync_metadata"`
	Version      int    `json:"version"`
}

type Roots struct {
	BookmarkBar Root `json:"bookmark_bar"`
	Other       Root `json:"other"`
	Synced      Root `json:"synced"`
}

type Root struct {
	Children     []BookmarkFolder `json:"children"`
	DateAdded    string           `json:"date_added"`
	DateLastUsed string           `json:"date_last_used"`
	DateModified string           `json:"date_modified"`
	GUID         string           `json:"guid"`
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Type         string           `json:"type"`
}

type BookmarkFolder struct {
	Children     []BookmarkFolder `json:"children,omitempty"`
	DateAdded    string           `json:"date_added"`
	DateLastUsed string           `json:"date_last_used"`
	DateModified string           `json:"date_modified,omitempty"`
	GUID         string           `json:"guid"`
	ID           string           `json:"id"`
	MetaInfo     MetaInfo         `json:"meta_info,omitempty"`
	Name         string           `json:"name"`
	Type         string           `json:"type"`
	URL          string           `json:"url,omitempty"`
}

type MetaInfo struct {
	PowerBookmarkMeta string `json:"power_bookmark_meta"`
}

func main() {
	historyRecords, err := GetHistoryRecords("")
	if err != nil {
		fmt.Println(err)

		return
	}

	historyRecords, err = ParseHistoryRecords(historyRecords)
	if err != nil {
		fmt.Println(err)

		return
	}

	historyRecords, err = DeleteDuplicates(historyRecords)
	if err != nil {
		fmt.Println(err)

		return
	}

	bookmarks, err := GetBookmarks("Bookmarks")
	if err != nil {
		fmt.Println(err)

		return
	}

	bookmarks, err = AddHistoryRecordsToBookmarks(historyRecords, bookmarks)
	if err != nil {
		fmt.Println(err)

		return
	}

	err = SaveBookmarks("Bookmarks-01", bookmarks)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("Done")
}

func GetHistoryRecords(filePath string) ([]HistoryRecord, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	historyRecords := make([]HistoryRecord, 0, HistoryRecordCount)

	err = json.Unmarshal(bytes, &historyRecords)
	if err != nil {
		return nil, err
	}

	return historyRecords, nil
}

func ParseHistoryRecords(historyRecords []HistoryRecord) ([]HistoryRecord, error) {
	handledHistoryRecords := make([]HistoryRecord, 0, len(historyRecords))

	for _, historyRecord := range historyRecords {
		if strings.Contains(historyRecord.URL, TargetURL) {
			handledHistoryRecords = append(handledHistoryRecords, historyRecord)
		}
	}

	return handledHistoryRecords, nil
}

func DeleteDuplicates(historyRecords []HistoryRecord) ([]HistoryRecord, error) {
	handledHistoryRecords := make([]HistoryRecord, 0, len(historyRecords))

	addedRecords := make(map[string]struct{}, len(historyRecords))

	for _, historyRecord := range historyRecords {
		_, ok := addedRecords[historyRecord.URL]
		if ok {
			continue
		}

		handledHistoryRecords = append(handledHistoryRecords, historyRecord)

		addedRecords[historyRecord.URL] = struct{}{}
	}

	return handledHistoryRecords, nil
}

func GetBookmarks(filePath string) (*Bookmarks, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	bookmarks := &Bookmarks{}

	err = json.Unmarshal(bytes, bookmarks)
	if err != nil {
		return nil, err
	}

	return bookmarks, nil
}

func AddHistoryRecordsToBookmarks(historyRecords []HistoryRecord, bookmarks *Bookmarks) (*Bookmarks, error) {
	bookmarkFolders := make([]BookmarkFolder, 0, len(historyRecords))

	ID := InitialID

	for _, historyRecord := range historyRecords {
		ID++

		bookmarkFolders = append(bookmarkFolders, BookmarkFolder{
			DateAdded:    "0",
			DateLastUsed: "0",
			GUID:         uuid.NewString(),
			ID:           strconv.Itoa(ID),
			Name:         historyRecord.Title,
			Type:         "url",
			URL:          historyRecord.URL,
		})
	}

	bookmarks.Roots.BookmarkBar.Children = append(bookmarks.Roots.BookmarkBar.Children, BookmarkFolder{
		Children:     bookmarkFolders,
		DateAdded:    "0",
		DateLastUsed: "0",
		GUID:         uuid.NewString(),
		ID:           strconv.Itoa(InitialID),
		Name:         FolderName,
		Type:         "folder",
	})

	return bookmarks, nil
}

func SaveBookmarks(filePath string, bookmarks *Bookmarks) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	encoder := json.NewEncoder(file)

	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "   ")

	return encoder.Encode(bookmarks)
}
