package docu

import "github.com/niiigoo/hawk/proto"

type Data struct {
	Service     *proto.Service
	Messages    map[string]Message
	Referenced  []string
	VersionName string
	VersionTime string
}

type Message struct {
	Name        string
	Description string
	Fields      []Field
}

type Field struct {
	Name        string
	Description string
	Type        string
	Optional    bool
	Options     []string
}
