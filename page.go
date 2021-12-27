package main

type Page struct {
	Parent         PageParent           `json:"parent,omitempty"`
	Object         *string              `json:"object,omitempty"`
	Id             *string              `json:"id,omitempty"`
	CreatedTime    *string              `json:"createdTime,omitempty"`
	LastEditedTime *string              `json:"lastEditedTime,omitempty"`
	Archived       *bool                `json:"archived,omitempty"`
	Icon           *FileObject          `json:"icon,omitempty"`
	Cover          *FileObject          `json:"cover,omitempty"`
	Properties     *map[string]Property `json:"properties,omitempty"`
	Url            *string              `json:"url,omitempty"`
}

type PageParent struct {
	Type       *string `json:"type,omitempty"`
	PageId     *string `json:"page_id,omitempty"`
	DatabaseId *string `json:"database_id,omitempty"`
}

func (p *Page) GetTitle() string {
	props := *p.Properties
	titleProp := props["Name"]
	titleArr := *titleProp.Title
	agg := ""
	for _, el := range titleArr {
		agg += *el.Text.Content
	}
	return agg
}
