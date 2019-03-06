package igg_reporter

import "testing"

func TestConvertMaptoRecordByReflect(t *testing.T) {
	r := make(map[string]interface{})
	r["snId"] = int64(1)
	r["userId"] = int64(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMapToRecordByReflect(r); record.SnId != 1 {
		t.FailNow()
	}
}
func TestConvertMapToRecord(t *testing.T) {
	r := make(map[string]interface{})
	r["snId"] = int64(1)
	r["userId"] = int64(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMapToRecord(r); record.SnId != 1 {
		t.FailNow()
	}
}
