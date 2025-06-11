package jsondsTemplate

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

func newUUIDv4() (string, error) {
	u := make([]byte, 16)
	if _, err := rand.Read(u); err != nil {
		return "", err
	}
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(u[0])<<24|uint32(u[1])<<16|uint32(u[2])<<8|uint32(u[3]),
		uint16(u[4])<<8|uint16(u[5]),
		uint16(u[6])<<8|uint16(u[7]),
		uint16(u[8])<<8|uint16(u[9]),
		hex.EncodeToString(u[10:16]),
	), nil
}
func addUUIDToItems(items []map[string]interface{}) error {
	for _, item := range items {
		id, err := newUUIDv4()
		if err != nil {
			return err
		}
		item["id"] = id
	}
	return nil
}
func TestAddUUIDToJSONItems(t *testing.T) {
	infile := "support_model.json"
	outfile := "support_model_with_uuid.json"
	data, err := ioutil.ReadFile(infile)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}
	var models []map[string]interface{}
	if err := json.Unmarshal(data, &models); err != nil {
		log.Fatalf("解析JSON失败: %v", err)
	}
	for _, item := range models {
		id, err := newUUIDv4()
		if err != nil {
			log.Fatalf("生成UUID失败: %v", err)
		}
		item["id"] = id
	}
	outBytes, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		log.Fatalf("序列化JSON失败: %v", err)
	}
	if err := ioutil.WriteFile(outfile, outBytes, 0644); err != nil {
		log.Fatalf("写入新文件失败: %v", err)
	}
	fmt.Printf("处理完成，新文件保存在：%s\n", outfile)
}
