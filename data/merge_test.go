package data

import (
	"reflect"
	"testing"
)

func TestMergeDataAtPath(t *testing.T) {
	t.Run("Merge at root", func(t *testing.T) {
		existing := map[string]interface{}{
			"a": "1",
		}
		newData := map[string]interface{}{
			"b": "2",
		}

		MergeDataAtPath(existing, "", newData)

		if existing["a"] != "1" {
			t.Errorf("Expected a=1, got %v", existing["a"])
		}
		if existing["b"] != "2" {
			t.Errorf("Expected b=2, got %v", existing["b"])
		}
	})

	t.Run("Merge at nested path", func(t *testing.T) {
		existing := map[string]interface{}{
			"room1": map[string]interface{}{
				"temp": "20",
			},
		}
		newData := map[string]interface{}{
			"humidity": "45",
		}

		MergeDataAtPath(existing, "room1", newData)

		room1, ok := existing["room1"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected room1 to be a map")
		}
		if room1["temp"] != "20" {
			t.Errorf("Expected temp=20, got %v", room1["temp"])
		}
		if room1["humidity"] != "45" {
			t.Errorf("Expected humidity=45, got %v", room1["humidity"])
		}
	})

	t.Run("Create nested path", func(t *testing.T) {
		existing := map[string]interface{}{}
		newData := map[string]interface{}{
			"temp": "22",
		}

		MergeDataAtPath(existing, "room1/sensors", newData)

		room1, ok := existing["room1"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected room1 to be a map")
		}
		sensors, ok := room1["sensors"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected sensors to be a map")
		}
		if sensors["temp"] != "22" {
			t.Errorf("Expected temp=22, got %v", sensors["temp"])
		}
	})

	t.Run("Empty path merges at root", func(t *testing.T) {
		existing := map[string]interface{}{
			"key1": "value1",
		}
		newData := map[string]interface{}{
			"key2": "value2",
		}

		MergeDataAtPath(existing, "", newData)

		expected := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		if !reflect.DeepEqual(existing, expected) {
			t.Errorf("got %v, want %v", existing, expected)
		}
	})

	t.Run("Deeply nested merge", func(t *testing.T) {
		existing := map[string]interface{}{
			"key1": map[string]interface{}{
				"subkey1": map[string]interface{}{
					"subsubkey1": "subsubvalue1",
				},
			},
		}
		newData := map[string]interface{}{
			"subsubkey2": "subsubvalue2",
		}

		MergeDataAtPath(existing, "key1/subkey1", newData)

		key1 := existing["key1"].(map[string]interface{})
		subkey1 := key1["subkey1"].(map[string]interface{})
		if subkey1["subsubkey1"] != "subsubvalue1" {
			t.Errorf("Expected subsubvalue1, got %v", subkey1["subsubkey1"])
		}
		if subkey1["subsubkey2"] != "subsubvalue2" {
			t.Errorf("Expected subsubvalue2, got %v", subkey1["subsubkey2"])
		}
	})

	t.Run("Override non-map value", func(t *testing.T) {
		existing := map[string]interface{}{
			"key1": "value1",
		}
		newData := map[string]interface{}{
			"subkey1": "subvalue1",
		}

		MergeDataAtPath(existing, "key1", newData)

		key1, ok := existing["key1"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected key1 to be a map after override")
		}
		if key1["subkey1"] != "subvalue1" {
			t.Errorf("Expected subvalue1, got %v", key1["subkey1"])
		}
	})

	t.Run("Create new nested map", func(t *testing.T) {
		existing := map[string]interface{}{
			"key1": "value1",
		}
		newData := map[string]interface{}{
			"subkey1": "subvalue1",
		}

		MergeDataAtPath(existing, "key2", newData)

		expected := map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{
				"subkey1": "subvalue1",
			},
		}
		if !reflect.DeepEqual(existing, expected) {
			t.Errorf("got %v, want %v", existing, expected)
		}
	})
}
