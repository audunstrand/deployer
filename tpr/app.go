package tpr

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
)

type AppSpec struct {
	Name string `json:"name"`
	Active bool   `json:"active"`
}

type App struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`

	Spec AppSpec `json:"spec"`
}

type AppList struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             unversioned.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}

// Required to satisfy Object interface
func (e *App) GetObjectKind() unversioned.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *App) GetObjectMeta() meta.Object {
	return &e.Metadata
}

// Required to satisfy Object interface
func (el *AppList) GetObjectKind() unversioned.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *AppList) GetListMeta() unversioned.List {
	return &el.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type AppListCopy AppList
type AppCopy App

func (e *App) UnmarshalJSON(data []byte) error {
	tmp := AppCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := App(tmp)
	*e = tmp2
	return nil
}

func (el *AppList) UnmarshalJSON(data []byte) error {
	tmp := AppListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := AppList(tmp)
	*el = tmp2
	return nil
}
