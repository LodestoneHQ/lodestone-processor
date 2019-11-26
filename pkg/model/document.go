package model

import "time"

type Document struct {
	ID string `json:"id"`

	Content string `json:"content"`

	Lodestone struct {
		Tags     []string `json:"tags"`
		Bookmark bool     `json:"bookmark"`
	} `json:"lodestone"`

	// File information/attributes
	File DocFile `json:"file"`

	// Document storage location (and thumbnail storage)
	Storage DocStorage `json:"storage"`

	// Document metadata extracted from document via tika
	Meta DocMeta `json:"meta"`
}

type DocFile struct {
	ContentType  string    `json:"content_type"`
	FileName     string    `json:"filename"`
	Extension    string    `json:"extension"`
	Filesize     int64     `json:"filesize"`
	IndexedChars int64     `json:"indexed_chars"`
	IndexedDate  time.Time `json:"indexed_date"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
	LastAccessed time.Time `json:"last_accessed"`
	Checksum     string    `json:"checksum"`

	Group string `json:"group"`
	Owner string `json:"owner"`
}

type DocStorage struct {
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
	ThumbBucket string `json:"thumb_bucket"`
	ThumbPath   string `json:"thumb_path"`
}

type DocMeta struct {
	Author       string    `json:"author"`
	Date         string    `json:"date"`
	Keywords     []string  `json:"keywords"`
	Title        string    `json:"title"`
	Language     string    `json:"language"`
	Format       string    `json:"format"`
	Identifier   string    `json:"identifier"`
	Contributor  string    `json:"contributor"`
	Modifier     string    `json:"modifier"`
	CreatorTool  string    `json:"creator_tool"`
	Publisher    string    `json:"publisher"`
	Relation     string    `json:"relation"`
	Rights       string    `json:"rights"`
	Source       string    `json:"source"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	Created      string    `json:"created"`
	PrintDate    time.Time `json:"print_date"`
	MetadataDate time.Time `json:"metadata_date"`
	Latitude     string    `json:"latitude"`
	Longitude    string    `json:"longitude"`
	Altitude     string    `json:"altitude"`
	Rating       byte      `json:"rating"`
	Comments     string    `json:"comments"`
}
