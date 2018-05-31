// Code generated by go-bindata.
// sources:
// data/1_add_users_table.down.sql
// data/1_add_users_table.up.sql
// data/2_add_roles_table.down.sql
// data/2_add_roles_table.up.sql
// DO NOT EDIT!

package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __1_add_users_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\xb6\xe6\x02\x04\x00\x00\xff\xff\xcf\x0c\x8a\x87\x12\x00\x00\x00")

func _1_add_users_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_add_users_tableDownSql,
		"1_add_users_table.down.sql",
	)
}

func _1_add_users_tableDownSql() (*asset, error) {
	bytes, err := _1_add_users_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_add_users_table.down.sql", size: 18, mode: os.FileMode(420), modTime: time.Unix(1519852554, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_add_users_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x8f\x31\x4b\x43\x31\x14\x85\xe7\xe4\x57\xdc\xf1\x3d\xc8\x50\xc5\xcd\xe9\xb6\xbd\xa5\xc1\xbc\xa4\xa6\x37\xd6\xba\x94\x40\x22\x0d\xf4\xf9\x24\x41\xfd\xfb\x62\x45\x1d\xba\x39\x9e\xc3\x77\x3e\x38\x0b\x4f\xc8\x04\x8c\x73\x43\xf0\xd6\x72\x6d\x9d\x14\x25\x41\x08\x7a\x09\x1b\xaf\x07\xf4\x7b\xb8\xa3\xbd\x92\x22\x8f\xb1\x9c\xe0\x01\xfd\x62\x8d\xbe\xbb\x9a\x5d\xdf\xf4\x60\x1d\x83\x0d\xc6\x40\xb0\xfa\x3e\x90\x92\xe2\x35\xb6\xf6\x31\xd5\x74\x38\xc6\x76\x04\xa6\x47\xfe\x85\x94\x14\x2f\x71\xcc\xe7\x52\x49\x11\x53\xca\x09\x58\x0f\xb4\x65\x1c\x36\xb0\xd3\xbc\x3e\x47\x78\x72\x96\xfe\xd4\x4b\x5a\x61\x30\x5f\x9a\x5d\xd7\x2b\x29\xc6\x29\x95\xe7\xf2\xaf\x69\x9d\x4e\x19\xb4\xe5\x4b\x62\xf6\x73\xf0\xf0\x9e\xeb\xb7\x7e\xee\x9c\x21\xb4\x97\xec\x0a\xcd\x96\x64\x7f\xfb\x19\x00\x00\xff\xff\x65\x74\xed\xf1\x3c\x01\x00\x00")

func _1_add_users_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_add_users_tableUpSql,
		"1_add_users_table.up.sql",
	)
}

func _1_add_users_tableUpSql() (*asset, error) {
	bytes, err := _1_add_users_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_add_users_table.up.sql", size: 316, mode: os.FileMode(420), modTime: time.Unix(1527364017, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __2_add_roles_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x28\xca\xcf\x49\x2d\xb6\x06\x04\x00\x00\xff\xff\xf9\xdd\xb1\x51\x11\x00\x00\x00")

func _2_add_roles_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__2_add_roles_tableDownSql,
		"2_add_roles_table.down.sql",
	)
}

func _2_add_roles_tableDownSql() (*asset, error) {
	bytes, err := _2_add_roles_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "2_add_roles_table.down.sql", size: 17, mode: os.FileMode(420), modTime: time.Unix(1527798209, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __2_add_roles_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x0e\x72\x75\x0c\x71\x55\x08\x71\x74\xf2\x71\x55\x28\xca\xcf\x49\x2d\xd6\xe0\xe2\x2c\x2d\x4e\x2d\x8a\xcf\x4c\x51\x08\x0d\xf5\x74\xd1\xe1\xe2\x04\x09\x2b\x84\x39\x06\x39\x7b\x38\x06\x69\x18\x1a\x18\x99\x68\xea\x70\x71\x06\x04\x79\xfa\x3a\x06\x45\x2a\x78\xbb\x46\x2a\x68\x40\x35\xe8\x80\x4d\xd0\xe4\xd2\xb4\x06\x04\x00\x00\xff\xff\xf5\x68\xf3\xff\x57\x00\x00\x00")

func _2_add_roles_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__2_add_roles_tableUpSql,
		"2_add_roles_table.up.sql",
	)
}

func _2_add_roles_tableUpSql() (*asset, error) {
	bytes, err := _2_add_roles_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "2_add_roles_table.up.sql", size: 87, mode: os.FileMode(420), modTime: time.Unix(1527798186, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"1_add_users_table.down.sql": _1_add_users_tableDownSql,
	"1_add_users_table.up.sql": _1_add_users_tableUpSql,
	"2_add_roles_table.down.sql": _2_add_roles_tableDownSql,
	"2_add_roles_table.up.sql": _2_add_roles_tableUpSql,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"1_add_users_table.down.sql": &bintree{_1_add_users_tableDownSql, map[string]*bintree{}},
	"1_add_users_table.up.sql": &bintree{_1_add_users_tableUpSql, map[string]*bintree{}},
	"2_add_roles_table.down.sql": &bintree{_2_add_roles_tableDownSql, map[string]*bintree{}},
	"2_add_roles_table.up.sql": &bintree{_2_add_roles_tableUpSql, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

