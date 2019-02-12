package igg_reporter

import "testing"

func TestConvertMaptoRecordByReflect(t *testing.T) {
	r := make(map[string]interface{})
	r["sn_id"] = int32(1)
	r["user_id"] = int32(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMaptoRecordByReflect(r); record.SnId != 1 {
		t.FailNow()
	}
}
func TestConvertMapToRecord(t *testing.T) {
	r := make(map[string]interface{})
	r["sn_id"] = int32(1)
	r["user_id"] = int32(1)
	r["website"] = "www.baidu.com"
	if record := ConvertMapToRecord(r); record.SnId != 1 {
		t.FailNow()
	}
}
