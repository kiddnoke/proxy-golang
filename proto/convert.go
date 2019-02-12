package igg_reporter

import (
	"reflect"
	"strings"
)

func ConvertMapToRecord(item map[string]interface{}) (record Record) {
	record.SnId = item["sn_id"].(int32)
	record.UserId = item["user_id"].(int32)
	record.DeviceId = item["device_id"].(string)
	record.Os = item["os"].(string)
	record.AppVersion = item["app_version"].(string)
	record.Ip = item["ip"].(string)
	record.TimeStamp = item["time_stamp"].(int64)
	record.Website = item["website"].(string)
	record.ConnectTime = item["connect_time"].(int32)
	record.Rate = item["rate"].(int32)
	record.Traffic = item["traffic"].(int32)
	record.CarrierOperator = item["carrier_operator"].(string)
	return
}
func ConvertMaptoRecordByReflect(item map[string]interface{}) (record Record) {
	record.Reset()
	TypeOfRecord := reflect.TypeOf(record)
	ValueOfRecord := reflect.ValueOf(&record).Elem()
	for i := 0; i < TypeOfRecord.NumField(); i++ {
		fieldType := TypeOfRecord.Field(i)
		filedValue := ValueOfRecord.Field(i)
		if !strings.Contains(fieldType.Tag.Get("json"), ",") {
			continue
		}
		split_array := strings.Split(fieldType.Tag.Get("json"), ",")
		key := split_array[0]

		if itemvalue, ok := item[key]; filedValue.CanSet() && ok {
			filedValue.Set(reflect.ValueOf(itemvalue))
		}
	}
	return
}
