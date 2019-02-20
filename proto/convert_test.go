package igg_reporter

import "testing"

func TestConvertMaptoRecordByReflect(t *testing.T) {
	r := make(map[string]interface{})
	r["sn_id"] = int64(1)
	r["user_id"] = int64(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMapToRecordByReflect(r); record.SnId != 1 {
		t.FailNow()
	}
}
func TestConvertMapToRecord(t *testing.T) {
	r := make(map[string]interface{})
	r["sn_id"] = int64(1)
	r["user_id"] = int64(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMapToRecord(r); record.SnId != 1 {
		t.FailNow()
	}
}
