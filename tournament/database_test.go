package tournament

import (
	"testing"
)

// import "code.google.com/p/gomock/gomock"

func checkParseSchema(t *testing.T, v string, majorE, minorE, patchE int) {
	if major, minor, patch, err := ParseSchemaVersion(v); err != nil {
		t.Error(err)
	} else if major != majorE || minor != minorE || patch != patchE {
		t.Errorf("expected %s to parse to [%d,%d,%d] actual [%d,%d,%d]", v, majorE, minorE, patchE, major, minor, patch)
	}
}

func checkParseSchemaError(t *testing.T, v string) {
	if major, minor, patch, err := ParseSchemaVersion(v); err == nil {
		t.Errorf("expected %s to parse fail with error actual [%d,%d,%d]", v, major, minor, patch)
	}
}
func TestParseSchemaVersion(t *testing.T) {
	checkParseSchema(t, "1.2.3", 1, 2, 3)
	checkParseSchema(t, "11.2.3", 11, 2, 3)
	checkParseSchema(t, "1.2.3_SNAPSHOT_2014-01-01", 1, 2, 3)
	checkParseSchemaError(t, "1.2")

}

func checkLessThan(t *testing.T, a, b string) {
	if !SchemaVersionLess(a, b) {
		t.Errorf("expected %s < %s", a, b)
	}
}

func checkNotLessThan(t *testing.T, a, b string) {
	if SchemaVersionLess(a, b) {
		t.Errorf("expected %s >= %s", a, b)
	}
}

func TestSchemaVersionOrdering(t *testing.T) {
	checkNotLessThan(t, "0.0.0", "0.0.0")
	checkNotLessThan(t, "0.0.1", "0.0.1")
	checkNotLessThan(t, "0.1.0", "0.1.0")
	checkLessThan(t, "0.0.0", "0.0.1")
	checkLessThan(t, "0.0.1", "0.1.0")
	checkLessThan(t, "0.1.0", "1.0.0")
	checkLessThan(t, "0.1.1", "1.0.0")
	checkLessThan(t, "0.0.1", "1.2.2")
	checkLessThan(t, "9.0.0", "11.0.0")
	checkNotLessThan(t, "11.0.0", "9.0.0")
}

func TestMigrateSchema(t *testing.T) {
	database, err := NewInMemoryDatabase()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if version, err := database.SchemaVersion(); err != nil {
		t.Error(err)
		t.FailNow()
	} else if version != ZeroVersion {
		t.Errorf("expected %s version, got %s", ZeroVersion, version)
		t.FailNow()
	}
	if err = database.MigrateSchema(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	maxVersion := SchemaVersion()
	if version, err := database.SchemaVersion(); err != nil {
		t.Error(err)
		t.FailNow()
	} else if version != maxVersion {
		t.Errorf("expected %s version, got %s", maxVersion, version)
		t.FailNow()
	}
}
