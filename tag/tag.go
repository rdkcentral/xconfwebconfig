package tag

import (
	"encoding/json"

	"github.com/rdkcentral/xconfwebconfig/util"
)

type Tag struct {
	Id      string   `json:"id"`
	Members util.Set `json:"members"`
	Updated int64    `json:"updated"`
}

func NewTagInf() interface{} {
	return &Tag{}
}

func (obj *Tag) Clone() (*Tag, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Tag), nil
}

// tagResp represents a response object for a tag. Used only for marshaling / unmarshaling
type tagResp struct {
	Id      string   `json:"id"`
	Members []string `json:"members"`
	Updated int64    `json:"updated"`
}

func (t *Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(tagResp{
		Id:      t.Id,
		Members: t.Members.ToSlice(),
		Updated: t.Updated,
	})
}

func (t *Tag) UnmarshalJSON(bbytes []byte) error {
	var tagResp tagResp
	err := json.Unmarshal(bbytes, &tagResp)
	if err != nil {
		return err
	}
	t.Id = tagResp.Id
	t.Updated = tagResp.Updated
	member := util.Set{}
	member.Add(tagResp.Members...)
	t.Members = member
	return nil
}
