package proto

import (
	"github.com/niiigoo/hawk/proto/io"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"strings"
)

type Parser interface {
	DetectFile(args ...string) (string, error)
	Parse(file string, comments ...bool) error
	ParseString(data string) error
	Definition() *Definition
	CreateFile(file, pgk, srv string) error
	CompileProto(file, out string, includes ...string) error
}

type service struct {
	file       string
	data       *io.Proto
	definition *Definition
}

func NewService() Parser {
	return &service{}
}

func (p *service) Definition() *Definition {
	return p.definition
}

func (p *service) DetectFile(args ...string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
			return dir + "/" + entry.Name(), nil
		}
	}

	return "", errors.New("no .proto file found")
}

func (p *service) Parse(file string, comments ...bool) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	p.file = file
	p.data, err = io.Parse(file, f, len(comments) > 0 && comments[0])
	if err != nil {
		return err
	}

	p.definition, err = DefinitionFromProto(p.data)

	return err
}

func (p *service) ParseString(data string) error {
	var err error
	p.data, err = io.ParseString("", data, false)
	if err != nil {
		return err
	}

	p.definition, err = DefinitionFromProto(p.data)

	return err
}

func (p *service) CreateFile(file, pkg, srv string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.Wrapf(err, "failed to open file '%s'", file)
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(`syntax = "proto3";

package ` + pkg + `;
option go_package = ".;` + pkg + `";

import "google/protobuf/descriptor.proto";
import "googleapis/google/api/annotations.proto";

extend google.protobuf.ServiceOptions {
  optional Config config = 10000;
}
extend google.protobuf.MethodOptions {
  optional bool httpCompress = 10001;
}

message Config {
  string HttpPrefix = 1;
  bool HttpCompress = 2;
  string WebSocketPath = 3;
  uint32 WebSocketMaxMessageSize = 4;
}

service ` + srv + ` {
	option (config) = {
		HttpPrefix: "/api/` + pkg + `"
		HttpCompress: false
	};
}
`)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%s'", file)
	}

	return nil
}

func (p *service) parseConfig() ProtocConfig {
	data, err := os.ReadFile("protoc.yaml")
	if err != nil {
		data, err = os.ReadFile("protoc.yml")
		if err != nil {
			return ProtocConfig{}
		}
	}

	var config ProtocConfig
	_ = yaml.Unmarshal(data, &config)
	return config
}

func (p *service) CompileProto(file, out string, imports ...string) error {
	config := p.parseConfig()
	if len(config.Imports) > 0 {
		imports = append(imports, config.Imports...)
	}

	args := []string{
		"--go-grpc_out=" + out,
		"--go_out=" + out,
	}
	for _, i := range imports {
		args = append(args, "-I="+os.ExpandEnv(i))
	}
	args = append(args, file)
	cmd := exec.Command("protoc", args...)
	err := cmd.Run()
	if err != nil {
		log.Info("Run: ", cmd.String())
		log.WithError(err).Error("failed to compile proto file")
	}
	return err
}
