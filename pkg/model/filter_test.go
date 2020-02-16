package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter_ValidPath(t *testing.T) {

	filter := Filter{
		Include: []string{"hello", "helloworld"},
		Exclude: []string{".DS_Store", ".*"},
	}

	assert.True(t, filter.ValidPath("subdirectory/hello"))

	assert.False(t, filter.ValidPath(".DS_Store"))
	assert.False(t, filter.ValidPath("subdirectory/.DS_Store"))
	assert.False(t, filter.ValidPath("subdirectory/.anything"))
	assert.False(t, filter.ValidPath(".anything"))

	assert.True(t, filter.ValidPath("hello/.DS_Store"))
}

func TestFilter_FromJson(t *testing.T) {

	bodyJson := `
{
  "include": [
    "*.doc",
    "*.docx",
    "*.xls",
    "*.xlsx",
    "*.ppt",
    "*.pptx",
    "*.pages",
    "*.numbers",
    "*.key",
    "*.pdf",
    "*.rtf",
    "*.md",
    "*.jpg",
    "*.jpeg",
    "*.png",
    "*.gif",
    "*.webp",
    "*.tiff",
    "*.tif",
    "*.html"
  ],
  "exclude": [
    ".*",

    ".DS_Store",
    ".AppleDouble",
    ".LSOverride",
    "._*",
    ".DocumentRevisions-V100",
    ".fseventsd",
    ".Spotlight-V100",
    ".TemporaryItems",
    ".Trashes",
    ".VolumeIcon.icns",
    ".com.apple.timemachine.donotpresent",
    ".AppleDB",
    ".AppleDesktop",
    "Network Trash Folder",
    "Temporary Items",
    ".apdisk",

    "Thumbs.db",
    "Thumbs.db:encryptable",
    "ehthumbs.db",
    "ehthumbs_vista.db",
    "*.stackdump",
    "Desktop.ini",
    "desktop.ini",
    "*.cab",
    "*.msi",
    "*.msix",
    "*.msm",
    "*.msp",
    "*.lnk"
  ]
}
`

	var filter Filter
	err := json.Unmarshal([]byte(bodyJson), &filter)
	assert.Nil(t, err)

	assert.True(t, filter.ValidPath("subdirectory/hello.png"))
	assert.True(t, filter.ValidPath("subdirectory/hello.jpeg"))
	assert.True(t, filter.ValidPath("subdirectory/hello.gif"))
	assert.True(t, filter.ValidPath("subdirectory/embedded.msiinside/inpath"))

	assert.False(t, filter.ValidPath(".DS_Store"))
	assert.False(t, filter.ValidPath("subdirectory/.DS_Store"))
	assert.False(t, filter.ValidPath("subdirectory/Thumbs.db"))
	assert.False(t, filter.ValidPath("helloworld/testing.cab"))
	assert.False(t, filter.ValidPath("subdirectory/.Trashes"))
}
