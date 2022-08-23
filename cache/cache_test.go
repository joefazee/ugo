package cache

import "testing"

func TestEncodeDecode(t *testing.T) {
	entry := Entry{}
	entry["foo"] = "bar"

	bys, err := encode(entry)
	if err != nil {
		t.Error(err)
	}

	out, err := decode(string(bys))
	if err != nil {
		t.Error(err)
	}

	if out["foo"] != "bar" {
		t.Errorf("expect decoded value to be bar; got %v", out)
	}
}
