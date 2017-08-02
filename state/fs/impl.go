// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type FileSystemStoreOptions struct {
	ClusterName string
	BasePath    string
}

type FileSystemStore struct {
	options      *FileSystemStoreOptions
	ClusterName  string
	BasePath     string
	AbsolutePath string
}

func NewFileSystemStore(o *FileSystemStoreOptions) *FileSystemStore {
	return &FileSystemStore{
		options:      o,
		ClusterName:  o.ClusterName,
		BasePath:     o.BasePath,
		AbsolutePath: fmt.Sprintf("%s/%s", o.BasePath, o.ClusterName),
	}
}

func (fs *FileSystemStore) Exists() bool {
	if _, err := os.Stat(fs.AbsolutePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fs *FileSystemStore) write(relativePath string, data []byte) error {
	fqn := fmt.Sprintf("%s/%s", fs.AbsolutePath, relativePath)
	os.MkdirAll(path.Dir(fqn), 0700)
	fo, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer fo.Close()
	_, err = io.Copy(fo, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileSystemStore) read(relativePath string) ([]byte, error) {
	fqn := fmt.Sprintf("%s/%s", fs.AbsolutePath, relativePath)
	bytes, err := ioutil.ReadFile(fqn)
	if err != nil {
		return []byte(""), err
	}
	return bytes, nil
}

func (fs *FileSystemStore) Commit(c *cluster.Cluster) error {
	if c == nil {
		return fmt.Errorf("Nil cluster spec!")
	}
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	fs.write(state.ClusterFile, bytes)
	return nil
}

func (fs *FileSystemStore) Rename(existingRelativePath, newRelativePath string) error {
	return os.Rename(existingRelativePath, newRelativePath)
}

func (fs *FileSystemStore) Destroy() error {
	logger.Warning("Removing path [%s]", fs.AbsolutePath)
	return os.RemoveAll(fs.AbsolutePath)
}

func (fs *FileSystemStore) GetCluster() (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	configBytes, err := fs.read(state.ClusterFile)
	if err != nil {
		return cluster, err
	}
	err = yaml.Unmarshal(configBytes, cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (fs *FileSystemStore) List() ([]string, error) {

	var state_list []string

	files, err := ioutil.ReadDir(fs.options.BasePath)
	if err != nil {
		return state_list, err
	}

	for _, file := range files {
		state_list = append(state_list, file.Name())
	}

	return state_list, nil
}
