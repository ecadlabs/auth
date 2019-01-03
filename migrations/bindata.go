// Code generated by go-bindata.
// sources:
// data/10_add_email_gen_column.down.sql
// data/10_add_email_gen_column.up.sql
// data/11_add_tenants_table.down.sql
// data/11_add_tenants_table.up.sql
// data/12_role_per_tenant.down.sql
// data/12_role_per_tenant.up.sql
// data/1_add_users_table.down.sql
// data/1_add_users_table.up.sql
// data/2_add_roles_table.down.sql
// data/2_add_roles_table.up.sql
// data/3_add_log_table.down.sql
// data/3_add_log_table.up.sql
// data/4_not_null.down.sql
// data/4_not_null.up.sql
// data/5_bootstrap.down.sql
// data/5_bootstrap.up.sql
// data/6_add_gen_column.down.sql
// data/6_add_gen_column.up.sql
// data/7_add_ip_column.down.sql
// data/7_add_ip_column.up.sql
// data/8_add_user_addr_columns.down.sql
// data/8_add_user_addr_columns.up.sql
// data/9_add_log_columns.down.sql
// data/9_add_log_columns.up.sql
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

var __10_add_email_gen_columnDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\x56\x70\x09\xf2\x0f\x50\x70\xf6\xf7\x09\xf5\xf5\x53\x48\xcd\x4d\xcc\xcc\x89\x4f\x4f\xcd\xb3\xe6\x02\x04\x00\x00\xff\xff\x15\xfc\x69\x9a\x29\x00\x00\x00")

func _10_add_email_gen_columnDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__10_add_email_gen_columnDownSql,
		"10_add_email_gen_column.down.sql",
	)
}

func _10_add_email_gen_columnDownSql() (*asset, error) {
	bytes, err := _10_add_email_gen_columnDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "10_add_email_gen_column.down.sql", size: 41, mode: os.FileMode(420), modTime: time.Unix(1532446858, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __10_add_email_gen_columnUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\x56\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\x48\xcd\x4d\xcc\xcc\x89\x4f\x4f\xcd\x53\xf0\xf4\x0b\x71\x75\x77\x0d\x52\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\x51\x70\x71\x75\x73\x0c\xf5\x09\x51\x30\xb0\xe6\x02\x04\x00\x00\xff\xff\x95\xf3\x7e\x91\x43\x00\x00\x00")

func _10_add_email_gen_columnUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__10_add_email_gen_columnUpSql,
		"10_add_email_gen_column.up.sql",
	)
}

func _10_add_email_gen_columnUpSql() (*asset, error) {
	bytes, err := _10_add_email_gen_columnUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "10_add_email_gen_column.up.sql", size: 67, mode: os.FileMode(420), modTime: time.Unix(1532446858, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __11_add_tenants_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xc8\x4d\xcd\x4d\x4a\x2d\x2a\xce\xc8\x2c\xb0\xe6\x42\x12\x2e\x49\xcd\x4b\xcc\x2b\x29\xb6\xe6\x82\x0a\x46\x06\xb8\x2a\x78\xba\x29\xb8\x46\x78\x06\x87\x04\x23\x69\x8a\x2f\xa9\x2c\x48\xd5\x41\x16\x28\x2e\x49\x2c\x29\x2d\xb6\x06\x04\x00\x00\xff\xff\x95\x9b\x6a\x0b\x63\x00\x00\x00")

func _11_add_tenants_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__11_add_tenants_tableDownSql,
		"11_add_tenants_table.down.sql",
	)
}

func _11_add_tenants_tableDownSql() (*asset, error) {
	bytes, err := _11_add_tenants_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "11_add_tenants_table.down.sql", size: 99, mode: os.FileMode(420), modTime: time.Unix(1546448952, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __11_add_tenants_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x94\x4f\x6f\xd3\x40\x10\xc5\xcf\xd9\x4f\xf1\x6e\xb6\x2b\x87\x0a\x89\x5b\xc4\x61\x9b\x4c\xa8\xc1\x5e\x47\xf6\x9a\x52\x2e\x91\xe9\x2e\x74\x45\x12\x47\xf6\x26\x88\x6f\x8f\xfc\x37\xa9\xd3\x02\xe2\xc2\xd1\xa3\x79\xbf\x9d\x99\x37\xe3\x1b\x7a\x17\x88\x19\x63\xf3\x84\xb8\x24\x48\x7e\x13\x12\xac\xde\xe5\x3b\x5b\xb9\x6c\x62\x14\xb2\x2c\x58\x60\x95\x04\x11\x4f\xee\xf1\x81\xee\xb1\xa0\x25\xcf\x42\x89\xc3\xc1\xa8\xf5\x37\xbd\xd3\x65\x6e\xf5\xfa\xf8\xc6\xf5\x7c\x36\xd9\xe5\x5b\x0d\x49\x9f\x24\x44\x2c\x21\xb2\x30\xf4\xd9\x24\x57\x4a\x2b\xc8\x20\xa2\x54\xf2\x68\x85\xbb\x40\xde\x36\x9f\xf8\x1c\x0b\x1a\x32\x07\xb2\x88\xef\x1a\xd8\xb6\x50\xe6\xab\xf9\x27\xe9\xbe\x2c\xac\x7e\xb0\x5a\xe1\x26\x8e\x43\xe2\xe2\x32\x75\xc9\xc3\x94\xea\xea\xca\x87\x47\x73\xfc\x63\x26\xf3\xce\xc6\x74\xbf\x22\x6c\xf5\xf6\x8b\x2e\xab\x47\xb3\x5f\xdb\x9f\x7b\x0d\x9e\x82\x44\x16\xc1\x75\x8a\x1f\x3b\x5d\x3a\x3e\x9c\x36\xc5\xf1\x66\x2f\x09\x2b\x9b\xdb\x43\x75\x26\xcd\x1f\xac\x39\xea\x5a\x6b\x76\x47\x63\xb5\x72\xbc\xb1\x3b\x27\xb9\xcb\x26\x87\x4a\x97\xeb\xde\xa5\x84\x96\x94\x90\x98\x53\x8a\x3a\x5e\xb9\x46\x79\x88\x05\xb2\xd5\xa2\x96\xcf\x79\x3a\xe7\x0b\xaa\x23\x0b\x0a\xe9\x14\xf1\xd9\xa4\xb5\xfc\x39\x52\xbf\x0c\x7f\xcf\xfa\x0f\x76\x8f\xbd\x18\x7f\x5f\x08\x7b\x6b\x9e\x6a\x3b\x3b\x2e\x23\x97\xfa\xde\x28\x06\xe0\xc9\x79\xb8\x9d\x25\x3e\x86\x99\x7a\xcd\xee\x4c\xa7\x08\x44\x4a\x89\x44\x20\x64\xdc\xcf\x15\x6e\x7d\x32\x3e\x86\x85\xf5\xf0\x91\x87\x19\xa5\x70\x9d\xb2\x28\xac\xe3\x43\x26\x19\x75\x80\xeb\x2b\x70\xa5\x90\x6f\x36\xd8\x97\xfa\x68\x8a\x43\xd5\x5a\x0d\x5b\xc0\x3e\x6a\xd4\x92\x0e\x8d\xab\xeb\xf1\x9b\xa7\xc6\x9e\x2f\x73\x3a\x45\x4a\x21\xcd\x25\x8c\xaa\x97\x72\xc8\x71\x9f\x84\x4f\xcb\xb2\x4c\xe2\x68\xe8\xe4\xee\x96\x12\x42\xf3\x07\xd8\x98\xef\x1a\x6d\xfd\x08\x83\x28\x90\x78\xed\xb5\xc9\x4d\xb5\x33\xc6\x5e\x1c\x85\x07\xd6\x3d\xa6\xb7\xb9\xd9\xd4\xef\x35\xc8\x17\xd5\xbf\x6f\xca\x1f\xef\x82\xd7\xe3\x1b\xd6\x2b\xa3\x90\x57\x18\xe9\xfa\xf0\x19\xa5\x3b\xea\xb3\x32\x10\xd2\x52\xe2\x7d\x1c\x88\xa1\x81\x58\x74\xd4\xb6\xf4\xb7\x03\xae\xee\xa0\xbe\xe3\x38\x8a\x02\x39\xfb\x15\x00\x00\xff\xff\x40\xed\x30\xed\x75\x05\x00\x00")

func _11_add_tenants_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__11_add_tenants_tableUpSql,
		"11_add_tenants_table.up.sql",
	)
}

func _11_add_tenants_tableUpSql() (*asset, error) {
	bytes, err := _11_add_tenants_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "11_add_tenants_table.up.sql", size: 1397, mode: os.FileMode(420), modTime: time.Unix(1546448132, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __12_role_per_tenantDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x90\x4d\x6a\xc3\x30\x10\x46\xf7\x73\x8a\x6f\x99\x40\x6f\xa0\x95\x22\x4d\xcb\x10\xfd\x84\xf1\x74\x91\x95\x29\x44\x8b\x40\x9b\x96\xc4\xbd\x7f\x89\x5d\x83\x31\x5e\xea\xf1\x10\xef\x9b\x03\xbf\x49\x71\x44\x3e\x19\x2b\xcc\x1f\x12\xe3\xfe\xfd\xd9\x1e\x88\x5a\x4f\x08\x35\xbd\xe7\x82\xa1\xdd\x3e\x6e\x43\x7f\xbd\x38\xa2\xa0\xec\x8d\x97\x6a\x3f\xb4\xaf\x1f\xec\x92\x1c\xff\xc1\xde\x11\x49\xe9\x58\x0d\x52\xac\x2e\xac\xdd\xef\xa3\xdd\xfb\xeb\xe5\x65\x64\x7b\xea\x38\x71\x30\x10\x00\x44\xe9\x4c\x4a\x30\xd4\x82\x95\x87\xf9\x39\x8a\x4f\x44\xaf\x5a\xf3\xf4\xb1\x23\x1a\x5b\x17\x45\x5b\x83\xa6\x4a\x52\x2e\x3e\x33\xe6\x2a\x87\xad\xed\x3e\x46\x9c\x54\xb2\xd7\x33\x8e\x7c\x5e\xe7\x3c\xaf\x50\x73\x16\x73\x7f\x01\x00\x00\xff\xff\xf0\xd1\x32\x15\x3f\x01\x00\x00")

func _12_role_per_tenantDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__12_role_per_tenantDownSql,
		"12_role_per_tenant.down.sql",
	)
}

func _12_role_per_tenantDownSql() (*asset, error) {
	bytes, err := _12_role_per_tenantDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "12_role_per_tenant.down.sql", size: 319, mode: os.FileMode(420), modTime: time.Unix(1546448917, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __12_role_per_tenantUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x91\xcb\x6a\xf3\x30\x14\x84\xf7\x7e\x8a\xd9\xe5\x42\x2e\xfc\x6b\x93\x85\x62\x9d\xfc\x55\x6b\xcb\x41\x96\x29\x59\x05\x53\x8b\x56\xc4\xb1\x83\xad\x40\xfb\xf6\x45\xb6\x9b\x96\xd2\x42\x77\x62\x74\x66\xbe\x91\xce\x96\xfe\x0b\x19\x06\x01\x8b\x35\x29\x68\xb6\x8d\x09\x6d\x53\x99\x0e\x8c\x73\x44\x69\x9c\x27\x12\xce\xd4\x45\xed\x8e\xb6\x44\x9e\x0b\x0e\x45\x3b\x52\x24\x23\xca\xc6\x9b\x6e\x6a\xcb\x19\x52\x89\x7c\xcf\x99\x26\x44\x2c\x8b\x18\x27\xaf\x70\x8a\xe9\x53\x09\x83\x60\xb9\xc4\x7a\x8e\xcc\x38\x14\x55\x05\xf3\x6a\x3b\x67\xeb\xe7\x91\x79\xad\x4b\xd3\xc2\xbd\x18\xb4\x4d\xe3\xc6\x74\xcc\xd7\xde\x35\x66\x0f\x83\x19\xe9\x2f\xad\x36\x98\x66\x14\x53\xa4\x61\x4b\xec\x54\x9a\x7c\xf4\xc2\xe3\x1d\x29\x42\x5d\x9c\x0d\x2a\x7b\x32\x98\xf8\xdc\x09\x98\xe4\xb8\xb4\x8d\x33\x4f\xce\x78\xbb\x56\x39\x21\x16\x89\xd0\xf8\x37\x0b\x83\xe0\x4f\xac\x91\xb1\xfa\xce\x8c\x69\xa7\x71\x9f\x0a\x89\x6b\x67\xda\xce\xff\x42\x7f\x58\x99\x73\x61\x2b\x6c\x6e\xc6\xbe\xd6\xd0\x70\x18\xe8\xe3\x7b\xe8\xca\x0b\x47\x5b\xce\x7e\x5c\x0d\x57\xe9\x1e\x51\x2a\x33\xad\x98\x90\x7a\x50\x8f\x97\x93\x79\x0b\x7f\x59\xe4\x5e\x89\x84\xa9\x03\x1e\xe8\x80\xe9\xed\x31\x0b\x8c\x98\x45\x3f\xea\x61\x51\x9a\x24\x42\x87\xef\x01\x00\x00\xff\xff\xe8\xd2\xda\xbb\x17\x02\x00\x00")

func _12_role_per_tenantUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__12_role_per_tenantUpSql,
		"12_role_per_tenant.up.sql",
	)
}

func _12_role_per_tenantUpSql() (*asset, error) {
	bytes, err := _12_role_per_tenantUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "12_role_per_tenant.up.sql", size: 535, mode: os.FileMode(420), modTime: time.Unix(1546449114, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
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

	info := bindataFileInfo{name: "1_add_users_table.down.sql", size: 18, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_add_users_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x8f\x41\x4b\xc4\x30\x10\x85\xcf\xc9\xaf\x98\x63\x0b\x39\xa8\x78\xf3\x34\xbb\x3b\xcb\x06\xd3\x64\xcd\x4e\x5c\xeb\xa5\x04\x12\x69\xc0\x5a\x69\x50\xff\xbe\x58\x41\x0f\xde\x3c\xbe\xc7\xf7\x3e\x78\x5b\x4f\xc8\x04\x8c\x1b\x43\xf0\x56\xf3\x52\x1b\x29\x4a\x82\x10\xf4\x0e\x8e\x5e\x77\xe8\x7b\xb8\xa5\x5e\x49\x91\xa7\x58\x9e\xe1\x1e\xfd\xf6\x80\xbe\xb9\xbc\xb8\xba\x6e\xc1\x3a\x06\x1b\x8c\x81\x60\xf5\x5d\x20\x25\xc5\x6b\xac\xf5\x63\x5e\xd2\x30\xc6\x3a\x02\xd3\x03\xff\x40\x4a\x8a\x97\x38\xe5\xb5\x54\x52\xc4\x94\x72\x02\xd6\x1d\x9d\x18\xbb\x23\x9c\x35\x1f\xd6\x08\x8f\xce\xd2\xaf\x7a\x47\x7b\x0c\xe6\x4b\x73\x6e\x5a\x25\xc5\x34\xa7\xf2\x54\xfe\x35\x5d\x2f\x0c\xef\x79\xf9\x16\x6c\x9c\x33\x84\xf6\x2f\xbf\x47\x73\x22\xd9\xde\x7c\x06\x00\x00\xff\xff\x1d\x0c\xf1\xf7\x1e\x01\x00\x00")

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

	info := bindataFileInfo{name: "1_add_users_table.up.sql", size: 286, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
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

	info := bindataFileInfo{name: "2_add_roles_table.down.sql", size: 17, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
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

	info := bindataFileInfo{name: "2_add_roles_table.up.sql", size: 87, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __3_add_log_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xc8\xc9\x4f\xb7\x06\x04\x00\x00\xff\xff\x5e\x0c\xb6\xd7\x0f\x00\x00\x00")

func _3_add_log_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__3_add_log_tableDownSql,
		"3_add_log_table.down.sql",
	)
}

func _3_add_log_tableDownSql() (*asset, error) {
	bytes, err := _3_add_log_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "3_add_log_table.down.sql", size: 15, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __3_add_log_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xc8\xb1\x0a\xc2\x30\x10\x06\xe0\xb9\x79\x8a\x7f\x6c\xa1\xa3\xb8\x38\x9d\xed\x49\x23\x69\x22\xe9\xc5\x82\x8b\x04\x1a\x8a\x20\x0a\x6d\xf4\xf9\x45\x27\xc7\xef\x6b\x3c\x93\x30\x84\xf6\x86\x71\x7f\xce\xa5\x2a\xf2\x0a\xd1\x3d\x0f\x42\xfd\x09\xa3\x96\xee\x47\x5c\x9c\x65\x58\x27\xb0\xc1\x18\xb4\x7c\xa0\x60\x04\xd6\x8d\x65\x55\xab\x22\xbd\xd3\x23\xe3\x4c\xbe\xe9\xc8\x97\xdb\xcd\xf7\x5e\x6b\x5a\xae\xb7\x09\x21\xe8\xb6\x56\x45\x8e\xcb\x9c\xf2\x5f\x4c\x31\x47\x1c\x07\x67\x55\xb5\xfb\x04\x00\x00\xff\xff\x68\xb3\x50\x13\x88\x00\x00\x00")

func _3_add_log_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__3_add_log_tableUpSql,
		"3_add_log_table.up.sql",
	)
}

func _3_add_log_tableUpSql() (*asset, error) {
	bytes, err := _3_add_log_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "3_add_log_table.up.sql", size: 136, mode: os.FileMode(420), modTime: time.Unix(1529430051, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __4_not_nullDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\x56\x80\x88\x38\xfb\xfb\x84\xfa\xfa\x29\xe4\x25\xe6\xa6\x2a\xb8\x04\xf9\x07\x28\xf8\xf9\x87\x28\xf8\x85\xfa\xf8\x58\x03\x02\x00\x00\xff\xff\xb2\xa0\xe0\x10\x32\x00\x00\x00")

func _4_not_nullDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__4_not_nullDownSql,
		"4_not_null.down.sql",
	)
}

func _4_not_nullDownSql() (*asset, error) {
	bytes, err := _4_not_nullDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "4_not_null.down.sql", size: 50, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __4_not_nullUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x72\x75\xf7\xf4\xb3\xe6\x0a\x0d\x70\x71\x0c\x71\x55\x28\x2d\x4e\x2d\x2a\x56\x08\x76\x0d\x51\xc8\x4b\xcc\x4d\x55\xb0\x55\x50\x57\x57\x08\xf7\x70\x0d\x72\x85\xf0\x3d\x83\x15\xfc\x42\x7d\x7c\xac\xb9\x1c\x7d\x42\x5c\x83\x14\x42\x1c\x9d\x7c\x60\x7a\x20\x22\xce\xfe\x3e\xa1\xbe\x7e\x10\xc5\x20\x53\xfc\xfc\x43\xc0\x3a\x74\x70\xc8\xbb\xb8\xba\x39\x86\xfa\x84\x28\xa8\xab\x5b\x73\x39\xfb\xfb\xfa\x7a\x86\x58\x03\x02\x00\x00\xff\xff\x51\xc4\xcf\x47\x91\x00\x00\x00")

func _4_not_nullUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__4_not_nullUpSql,
		"4_not_null.up.sql",
	)
}

func _4_not_nullUpSql() (*asset, error) {
	bytes, err := _4_not_nullUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "4_not_null.up.sql", size: 145, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __5_bootstrapDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\xca\xcf\x2f\x29\x2e\x29\x4a\x2c\xb0\x06\x04\x00\x00\xff\xff\xf2\x59\xf5\x1a\x15\x00\x00\x00")

func _5_bootstrapDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__5_bootstrapDownSql,
		"5_bootstrap.down.sql",
	)
}

func _5_bootstrapDownSql() (*asset, error) {
	bytes, err := _5_bootstrapDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "5_bootstrap.down.sql", size: 21, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __5_bootstrapUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x44\xcb\x31\x0e\xc2\x30\x0c\x05\xd0\x19\x9f\xe2\x8f\xed\x19\x32\x39\xc5\x45\x91\x5c\x5b\x6a\x1c\xf6\x30\x57\x2a\x82\xaa\xe7\x47\xb0\xb0\xbf\x97\xe5\x56\x2c\x11\x4d\xab\x70\x08\x82\xb3\x0a\x1e\xfb\x7e\xbc\x8f\x57\x7f\x0e\x74\x39\xfb\x86\xec\xae\xc2\x06\xf3\x80\x35\x55\x5c\x65\xe6\xa6\x81\x99\xb5\x0a\x8d\x89\xa8\x58\x95\x35\x50\x2c\xfc\xdf\x31\x9c\x7d\x1b\x71\x67\x6d\x52\x31\xfc\xf4\x17\x4f\xbe\x2c\x25\xd2\x27\x00\x00\xff\xff\xbd\x0c\x52\x7a\x7c\x00\x00\x00")

func _5_bootstrapUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__5_bootstrapUpSql,
		"5_bootstrap.up.sql",
	)
}

func _5_bootstrapUpSql() (*asset, error) {
	bytes, err := _5_bootstrapUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "5_bootstrap.up.sql", size: 124, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __6_add_gen_columnDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\x56\x70\x09\xf2\x0f\x50\x70\xf6\xf7\x09\xf5\xf5\x53\x28\x48\x2c\x2e\x2e\xcf\x2f\x4a\x89\x4f\x4f\xcd\xb3\x06\x04\x00\x00\xff\xff\xf7\x09\xd5\xe9\x2b\x00\x00\x00")

func _6_add_gen_columnDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__6_add_gen_columnDownSql,
		"6_add_gen_column.down.sql",
	)
}

func _6_add_gen_columnDownSql() (*asset, error) {
	bytes, err := _6_add_gen_columnDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "6_add_gen_column.down.sql", size: 43, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __6_add_gen_columnUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x04\xc0\x31\x0a\x80\x30\x0c\x05\xd0\xab\xfc\x23\xb8\x3b\x45\x1b\x45\x88\x29\x94\x64\x16\xc1\xe2\xa6\xd2\x20\x5e\xdf\x47\x62\x5c\x60\x34\x08\xe3\x8d\xda\x02\x94\x12\xc6\x2c\xbe\x2a\x9e\x3d\xe2\xbb\xdb\xb1\x9d\xf5\xc2\xa2\xc6\x33\x17\x68\x36\xa8\x8b\x20\xf1\x44\x2e\x86\xae\xff\x03\x00\x00\xff\xff\x36\x26\x6d\xf2\x45\x00\x00\x00")

func _6_add_gen_columnUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__6_add_gen_columnUpSql,
		"6_add_gen_column.up.sql",
	)
}

func _6_add_gen_columnUpSql() (*asset, error) {
	bytes, err := _6_add_gen_columnUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "6_add_gen_column.up.sql", size: 69, mode: os.FileMode(420), modTime: time.Unix(1529711024, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __7_add_ip_columnDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\xc8\xc9\x4f\x57\x70\x09\xf2\x0f\x50\x70\xf6\xf7\x09\xf5\xf5\x53\x48\x4c\x49\x29\xb2\xe6\x02\x04\x00\x00\xff\xff\x85\x4a\x91\x9c\x22\x00\x00\x00")

func _7_add_ip_columnDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__7_add_ip_columnDownSql,
		"7_add_ip_column.down.sql",
	)
}

func _7_add_ip_columnDownSql() (*asset, error) {
	bytes, err := _7_add_ip_columnDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "7_add_ip_column.down.sql", size: 34, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __7_add_ip_columnUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\xc8\xc9\x4f\x57\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\x48\x4c\x49\x29\x52\x08\x73\x0c\x72\xf6\x70\x0c\xd2\x30\x33\xd1\xb4\x06\x04\x00\x00\xff\xff\x61\xf7\x7c\xff\x2c\x00\x00\x00")

func _7_add_ip_columnUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__7_add_ip_columnUpSql,
		"7_add_ip_column.up.sql",
	)
}

func _7_add_ip_columnUpSql() (*asset, error) {
	bytes, err := _7_add_ip_columnUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "7_add_ip_column.up.sql", size: 44, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __8_add_user_addr_columnsDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\xe6\xe2\x74\x09\xf2\x0f\x50\x70\xf6\xf7\x09\xf5\xf5\x53\xc8\x49\x2c\x2e\x89\xcf\xc9\x4f\xcf\xcc\x8b\x4f\x4c\x49\x29\xd2\xc1\x29\x5b\x52\x8c\x4d\xae\x28\x35\xad\x28\xb5\x38\x03\xa7\x5e\x98\x7c\x49\xb1\x35\x17\x20\x00\x00\xff\xff\xb1\xe2\xea\x62\x8a\x00\x00\x00")

func _8_add_user_addr_columnsDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__8_add_user_addr_columnsDownSql,
		"8_add_user_addr_columns.down.sql",
	)
}

func _8_add_user_addr_columnsDownSql() (*asset, error) {
	bytes, err := _8_add_user_addr_columnsDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "8_add_user_addr_columns.down.sql", size: 138, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __8_add_user_addr_columnsUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\xe6\xe2\x74\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\xc8\xc9\x4f\xcf\xcc\x8b\x4f\x4c\x49\x29\x52\x08\x73\x0c\x72\xf6\x70\x0c\xd2\x30\x33\xd1\x54\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\x51\x70\x71\x75\x73\x0c\xf5\x09\x51\x50\x57\xd7\xc1\xa2\xaf\xa4\x58\x21\xc4\xd3\xd7\x35\x38\xc4\xd1\x37\x40\x21\xdc\x33\xc4\x03\xcc\x55\x88\xf2\xf7\x73\xc5\x62\x44\x6a\x41\x7e\x72\x06\x9a\x39\x45\xa9\x69\x45\xa9\xc5\x19\x64\xb8\x00\xa6\x93\x1c\x37\x58\x73\x01\x02\x00\x00\xff\xff\x07\xaf\xae\x99\x16\x01\x00\x00")

func _8_add_user_addr_columnsUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__8_add_user_addr_columnsUpSql,
		"8_add_user_addr_columns.up.sql",
	)
}

func _8_add_user_addr_columnsUpSql() (*asset, error) {
	bytes, err := _8_add_user_addr_columnsUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "8_add_user_addr_columns.up.sql", size: 278, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __9_add_log_columnsDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\xcb\x4d\x0a\xc2\x30\x10\xc5\xf1\x75\xe7\x14\x43\xd7\xf5\x04\x59\xb5\x35\x48\x21\x99\x48\x4d\xc1\x9d\x14\x26\x84\x80\x5a\xc9\x87\xe7\x17\xaa\x1b\x15\xba\x7c\x7f\x7e\xaf\x93\x87\x81\x04\x40\xab\xac\x1c\xd1\xb6\x9d\x92\x78\x5d\x3c\x54\xfb\xd1\x1c\xb1\x37\x6a\xd2\x84\x81\x9b\xef\x70\x4b\xbe\x81\xea\xfd\xf9\x24\xf7\x74\xf7\x8c\x2b\x22\x63\x91\x26\xa5\x7e\xc9\xcc\x1c\xb7\x45\x49\x2e\x5e\x02\x6f\xa3\x3c\x47\xef\xf2\x1f\x13\x00\xeb\x96\x67\x2b\xe9\x34\x18\xc2\xba\x94\xc0\xbb\x25\xa5\x47\x2d\x00\x7a\xa3\xf5\x60\xc5\x2b\x00\x00\xff\xff\x73\xb6\x3d\x95\xf1\x00\x00\x00")

func _9_add_log_columnsDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__9_add_log_columnsDownSql,
		"9_add_log_columns.down.sql",
	)
}

func _9_add_log_columnsDownSql() (*asset, error) {
	bytes, err := _9_add_log_columnsDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "9_add_log_columns.down.sql", size: 241, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __9_add_log_columnsUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x91\xcd\x6e\x82\x40\x14\x85\xd7\x9d\xa7\x38\x71\xa3\x26\x76\xd7\x1d\x71\x31\xca\xb5\x9d\x74\x18\x0c\xdc\x49\x71\x45\x48\x98\x10\x12\x0a\x0d\x3f\x3e\x7f\x03\xb5\xda\x1a\x75\x39\x73\xbf\xef\x64\xce\x9d\x0d\xbd\x2a\xe3\x09\xb1\x8d\x48\x32\x81\x12\x26\x13\xab\xd0\x40\xed\x60\x42\x06\x25\x2a\xe6\x18\xb3\x61\x28\xf3\xe7\xa6\xeb\xbe\x66\x9e\x10\x52\x33\x45\x60\xb9\xd1\x84\xaa\x29\xc4\x93\xf4\x7d\x6c\x43\x6d\x03\x83\x32\x87\xb5\xca\xc7\x3e\x52\x81\x8c\x0e\x78\xa7\x03\x7c\xda\x49\xab\x19\x63\x48\x5a\xb8\xda\xb5\x59\xef\xd2\xe3\xcb\x62\xb9\xfa\xe7\x7e\x76\x05\x98\x12\xf6\x84\xb0\x7b\x7f\x7c\x4f\xd5\x14\x88\x89\x31\x74\xae\x4d\xcb\x1c\xeb\x9f\x8c\xba\xac\x16\x4b\x7c\xbc\x51\x44\xe7\x91\x8a\x61\xac\xd6\xde\xb5\xda\x67\x6d\xe1\xfa\x3b\xf2\x65\x78\x4f\xcf\xf2\xbc\xc5\x1a\xf3\xf9\xc9\x98\xce\x67\xf8\xc6\x2a\xa6\x8b\x53\x21\x77\x74\x75\x3f\xc5\x8c\xbb\x1c\x95\xd5\x15\x31\xc5\x3d\x02\x7e\xeb\x3d\x62\x2e\x2d\xfe\x52\xe3\xaf\x86\x41\xa0\xd8\x13\xdf\x01\x00\x00\xff\xff\xb3\x47\x29\x12\xe6\x01\x00\x00")

func _9_add_log_columnsUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__9_add_log_columnsUpSql,
		"9_add_log_columns.up.sql",
	)
}

func _9_add_log_columnsUpSql() (*asset, error) {
	bytes, err := _9_add_log_columnsUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "9_add_log_columns.up.sql", size: 486, mode: os.FileMode(420), modTime: time.Unix(1530306711, 0)}
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
	"10_add_email_gen_column.down.sql": _10_add_email_gen_columnDownSql,
	"10_add_email_gen_column.up.sql": _10_add_email_gen_columnUpSql,
	"11_add_tenants_table.down.sql": _11_add_tenants_tableDownSql,
	"11_add_tenants_table.up.sql": _11_add_tenants_tableUpSql,
	"12_role_per_tenant.down.sql": _12_role_per_tenantDownSql,
	"12_role_per_tenant.up.sql": _12_role_per_tenantUpSql,
	"1_add_users_table.down.sql": _1_add_users_tableDownSql,
	"1_add_users_table.up.sql": _1_add_users_tableUpSql,
	"2_add_roles_table.down.sql": _2_add_roles_tableDownSql,
	"2_add_roles_table.up.sql": _2_add_roles_tableUpSql,
	"3_add_log_table.down.sql": _3_add_log_tableDownSql,
	"3_add_log_table.up.sql": _3_add_log_tableUpSql,
	"4_not_null.down.sql": _4_not_nullDownSql,
	"4_not_null.up.sql": _4_not_nullUpSql,
	"5_bootstrap.down.sql": _5_bootstrapDownSql,
	"5_bootstrap.up.sql": _5_bootstrapUpSql,
	"6_add_gen_column.down.sql": _6_add_gen_columnDownSql,
	"6_add_gen_column.up.sql": _6_add_gen_columnUpSql,
	"7_add_ip_column.down.sql": _7_add_ip_columnDownSql,
	"7_add_ip_column.up.sql": _7_add_ip_columnUpSql,
	"8_add_user_addr_columns.down.sql": _8_add_user_addr_columnsDownSql,
	"8_add_user_addr_columns.up.sql": _8_add_user_addr_columnsUpSql,
	"9_add_log_columns.down.sql": _9_add_log_columnsDownSql,
	"9_add_log_columns.up.sql": _9_add_log_columnsUpSql,
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
	"10_add_email_gen_column.down.sql": &bintree{_10_add_email_gen_columnDownSql, map[string]*bintree{}},
	"10_add_email_gen_column.up.sql": &bintree{_10_add_email_gen_columnUpSql, map[string]*bintree{}},
	"11_add_tenants_table.down.sql": &bintree{_11_add_tenants_tableDownSql, map[string]*bintree{}},
	"11_add_tenants_table.up.sql": &bintree{_11_add_tenants_tableUpSql, map[string]*bintree{}},
	"12_role_per_tenant.down.sql": &bintree{_12_role_per_tenantDownSql, map[string]*bintree{}},
	"12_role_per_tenant.up.sql": &bintree{_12_role_per_tenantUpSql, map[string]*bintree{}},
	"1_add_users_table.down.sql": &bintree{_1_add_users_tableDownSql, map[string]*bintree{}},
	"1_add_users_table.up.sql": &bintree{_1_add_users_tableUpSql, map[string]*bintree{}},
	"2_add_roles_table.down.sql": &bintree{_2_add_roles_tableDownSql, map[string]*bintree{}},
	"2_add_roles_table.up.sql": &bintree{_2_add_roles_tableUpSql, map[string]*bintree{}},
	"3_add_log_table.down.sql": &bintree{_3_add_log_tableDownSql, map[string]*bintree{}},
	"3_add_log_table.up.sql": &bintree{_3_add_log_tableUpSql, map[string]*bintree{}},
	"4_not_null.down.sql": &bintree{_4_not_nullDownSql, map[string]*bintree{}},
	"4_not_null.up.sql": &bintree{_4_not_nullUpSql, map[string]*bintree{}},
	"5_bootstrap.down.sql": &bintree{_5_bootstrapDownSql, map[string]*bintree{}},
	"5_bootstrap.up.sql": &bintree{_5_bootstrapUpSql, map[string]*bintree{}},
	"6_add_gen_column.down.sql": &bintree{_6_add_gen_columnDownSql, map[string]*bintree{}},
	"6_add_gen_column.up.sql": &bintree{_6_add_gen_columnUpSql, map[string]*bintree{}},
	"7_add_ip_column.down.sql": &bintree{_7_add_ip_columnDownSql, map[string]*bintree{}},
	"7_add_ip_column.up.sql": &bintree{_7_add_ip_columnUpSql, map[string]*bintree{}},
	"8_add_user_addr_columns.down.sql": &bintree{_8_add_user_addr_columnsDownSql, map[string]*bintree{}},
	"8_add_user_addr_columns.up.sql": &bintree{_8_add_user_addr_columnsUpSql, map[string]*bintree{}},
	"9_add_log_columns.down.sql": &bintree{_9_add_log_columnsDownSql, map[string]*bintree{}},
	"9_add_log_columns.up.sql": &bintree{_9_add_log_columnsUpSql, map[string]*bintree{}},
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

