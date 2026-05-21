package bson

import (
	"testing"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type bsonPlayer struct {
	Id    string
	Name  string
	Lv    int32
	Info  bsonInfo
	Items []int32
}

type bsonInfo struct {
	Vip  int32
	Desc string
}

func newTestDoc(t *testing.T) Document {
	t.Helper()
	player := &bsonPlayer{
		Id:    "1",
		Name:  "hwc",
		Lv:    100,
		Items: []int32{1, 2},
	}
	doc, err := Marshal(player)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	return doc
}

func makeTestDoc() Document {
	player := &bsonPlayer{
		Id:    "1",
		Name:  "hwc",
		Lv:    100,
		Items: []int32{1, 2},
	}
	doc, _ := Marshal(player)
	return doc
}

func TestMarshalAndGet(t *testing.T) {
	doc := newTestDoc(t)

	if got := doc.GetString("name"); got != "hwc" {
		t.Errorf("GetString(name) = %q, want %q", got, "hwc")
	}
	if got := doc.GetInt32("lv"); got != 100 {
		t.Errorf("GetInt32(lv) = %d, want %d", got, 100)
	}
}

func TestSetAndGetDotNotation(t *testing.T) {
	doc := newTestDoc(t)

	if err := doc.Set("info.vip", 100); err != nil {
		t.Fatalf("Set(info.vip) failed: %v", err)
	}
	if got := doc.GetInt32("info.vip"); got != 100 {
		t.Errorf("GetInt32(info.vip) = %d, want %d", got, 100)
	}
}

func TestSetArrayElement(t *testing.T) {
	doc := newTestDoc(t)

	if err := doc.Set("items.5", 50); err != nil {
		t.Fatalf("Set(items.5) failed: %v", err)
	}
	if got := doc.GetInt32("items.5"); got != 50 {
		t.Errorf("GetInt32(items.5) = %d, want %d", got, 50)
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	doc := newTestDoc(t)
	if err := doc.Set("info.vip", 42); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	result := &bsonPlayer{}
	if err := doc.Unmarshal(result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if result.Name != "hwc" {
		t.Errorf("Name = %q, want %q", result.Name, "hwc")
	}
	if result.Info.Vip != 42 {
		t.Errorf("Info.Vip = %d, want %d", result.Info.Vip, 42)
	}
}

func TestDocumentResetClearsOldKeys(t *testing.T) {
	doc := New()
	_ = doc.Set("old_key", "old_value")

	doc2 := newTestDoc(t)
	rawBytes := doc2.Raw(nil)
	if err := doc.Reset(rawBytes); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if doc.Has("old_key") {
		t.Error("Reset should have cleared old_key, but it still exists")
	}
}

func TestIsNil(t *testing.T) {
	ele, err := NewElement(TypeNull, nil)
	if err != nil {
		t.Fatalf("NewElement failed: %v", err)
	}
	if !ele.IsNil() {
		t.Error("Element with TypeNull should be nil")
	}

	ele2, err := NewElement(TypeInt32, []byte{1, 0, 0, 0})
	if err != nil {
		t.Fatalf("NewElement failed: %v", err)
	}
	if ele2.IsNil() {
		t.Error("Element with TypeInt32 should not be nil")
	}
}

func TestGetInt32FromInt64(t *testing.T) {
	doc := New()
	if err := doc.Set("val", int64(42)); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if got := doc.GetInt32("val"); got != 42 {
		t.Errorf("GetInt32 from int64 = %d, want %d", got, 42)
	}
}

func TestGetInt64FromInt32(t *testing.T) {
	doc := New()
	if err := doc.Set("val", int32(42)); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if got := doc.GetInt64("val"); got != 42 {
		t.Errorf("GetInt64 from int32 = %d, want %d", got, 42)
	}
}

func TestGetFloatFromInt32(t *testing.T) {
	doc := New()
	if err := doc.Set("val", int32(100)); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if got := doc.GetFloat("val"); got != 100.0 {
		t.Errorf("GetFloat from int32 = %f, want %f", got, 100.0)
	}
}

func TestGetFloatFromDouble(t *testing.T) {
	doc := New()
	if err := doc.Set("val", 3.14); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if got := doc.GetFloat("val"); got != 3.14 {
		t.Errorf("GetFloat from double = %f, want %f", got, 3.14)
	}
}

func TestUnset(t *testing.T) {
	doc := newTestDoc(t)
	if err := doc.Unset("name"); err != nil {
		t.Fatalf("Unset failed: %v", err)
	}
	if doc.Has("name") {
		t.Error("key 'name' should not exist after Unset")
	}
}

func BenchmarkBsonSet(b *testing.B) {
	doc := makeTestDoc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Set("info.vip", 100)
	}
}

func BenchmarkBsonGet(b *testing.B) {
	doc := makeTestDoc()
	_ = doc.Set("info.vip", int32(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.GetInt32("info.vip")
	}
}

func BenchmarkBytes(b *testing.B) {
	doc := makeTestDoc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Raw(nil)
	}
}

func BenchmarkMetaGet(b *testing.B) {
	doc := makeTestDoc()
	raw := doc.Raw(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := bsoncore.Document(raw)
		v := d.Lookup("info.vip")
		_, _ = v.Int32OK()
	}
}
