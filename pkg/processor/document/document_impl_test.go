package document

import (
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDocumentProcessor_Process(t *testing.T) {

	//setup
	proc := DocumentProcessor{}
	docModel := model.Document{}

	//test
	err := proc.parseTikaMetadata(`{
		"Accessibility": "structured; tagged",
		"Author": "SE:W:CAR:MP",
		"Content-Type": "application/pdf",
		"Creation-Date": "2018-10-24T18:27:15Z",
		"Form fields": "fillable",
		"Keywords": "Fillable",
		"Last-Modified": "2018-10-24T18:27:15Z",
		"Last-Save-Date": "2018-10-24T18:27:15Z",
		"X-Parsed-By": ["org.apache.tika.parser.DefaultParser", "org.apache.tika.parser.pdf.PDFParser"],
		"access_permission:assemble_document": "true",
		"access_permission:can_modify": "true",
		"access_permission:can_print": "true",
		"access_permission:can_print_degraded": "true",
		"access_permission:extract_content": "true",
		"access_permission:extract_for_accessibility": "true",
		"access_permission:fill_in_form": "true",
		"access_permission:modify_annotations": "true",
		"cp:subject": "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
		"created": "2018-10-24T18:27:15Z",
		"creator": "SE:W:CAR:MP",
		"date": "2018-10-24T18:27:15Z",
		"dc:creator": "SE:W:CAR:MP",
		"dc:description": "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
		"dc:format": ["application/pdf; version\u003d1.7", "application/pdf; version\u003d\"1.7 Adobe Extension Level 5\""],
		"dc:subject": "Fillable",
		"dc:title": "2018 Form 4868",
		"dcterms:created": "2018-10-24T18:27:15Z",
		"dcterms:modified": "2018-10-24T18:27:15Z",
		"description": "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
		"language": "en",
		"meta:author": "SE:W:CAR:MP",
		"meta:creation-date": "2018-10-24T18:27:15Z",
		"meta:keyword": "Fillable",
		"meta:save-date": "2018-10-24T18:27:15Z",
		"modified": "2018-10-24T18:27:15Z",
		"pdf:PDFExtensionVersion": "1.7 Adobe Extension Level 5",
		"pdf:PDFVersion": "1.7",
		"pdf:charsPerPage": ["5339", "5284", "6173", "4918"],
		"pdf:docinfo:created": "2018-10-24T18:27:15Z",
		"pdf:docinfo:creator": "SE:W:CAR:MP",
		"pdf:docinfo:creator_tool": "Adobe LiveCycle Designer ES 9.0",
		"pdf:docinfo:custom:Accessibility": "structured; tagged",
		"pdf:docinfo:custom:Form fields": "fillable",
		"pdf:docinfo:keywords": "Fillable",
		"pdf:docinfo:modified": "2018-10-24T18:27:15Z",
		"pdf:docinfo:producer": "Adobe LiveCycle Designer ES 9.0",
		"pdf:docinfo:subject": "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
		"pdf:docinfo:title": "2018 Form 4868",
		"pdf:encrypted": "false",
		"pdf:unmappedUnicodeCharsPerPage": ["0", "0", "0", "0"],
		"producer": "Adobe LiveCycle Designer ES 9.0",
		"subject": "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
		"title": "2018 Form 4868",
		"xmp:CreatorTool": "Adobe LiveCycle Designer ES 9.0",
		"xmpMM:DocumentID": "uuid:140c797f-30d2-4145-a45d-56b02215393e",
		"xmpTPg:NPages": "4"
	}`, &docModel)

	//assert
	require.NoError(t, err)
	require.Equal(t, model.Document{
		ID:      "",
		Content: "",
		Lodestone: model.DocLodestone{
			Tags:     []string(nil),
			Bookmark: false,
		},
		File: model.DocFile{
			ContentType:  "application/pdf",
			FileName:     "",
			Extension:    "",
			Filesize:     0,
			IndexedChars: 0,
			Checksum:     "",
			Group:        "",
			Owner:        "",
		},
		Storage: model.DocStorage{
			Bucket:      "",
			Path:        "",
			ThumbBucket: "",
			ThumbPath:   "",
		},
		Meta: model.DocMeta{
			Author:      "SE:W:CAR:MP",
			Date:        time.Time{},
			CreatedDate: time.Date(2018, time.October, 24, 18, 27, 15, 0, time.UTC), //"2018-10-24T18:27:15Z"),
			SavedDate:   time.Date(2018, time.October, 24, 18, 27, 15, 0, time.UTC),
			Keywords:    []string{"Fillable"},
			Title:       "2018 Form 4868",
			Language:    "en",
			Format:      "",
			Identifier:  "",
			Contributor: "",
			Modifier:    "",
			CreatorTool: "Adobe LiveCycle Designer ES 9.0",
			Publisher:   "",
			Relation:    "",
			Rights:      "",
			Source:      "",
			Type:        "",
			Description: "Application for Automatic Extension of Time To File U.S. Individual Income Tax Return",
			Latitude:    "",
			Longitude:   "",
			Altitude:    "",
			Rating:      0x0,
			Comments:    "",
			Pages:       "4",
		},
	}, docModel)
	//require.Implements(t, (*Interface)(nil), config, "should implement the config interface")
}
