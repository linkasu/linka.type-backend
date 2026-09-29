package migrate

import "testing"

func TestPlanIsIdempotentAndRejectsChecksumChanges(t *testing.T) {
	migrations := []Migration{{Version: 1, Checksum: "one"}, {Version: 2, Checksum: "two"}}
	plan, err := Plan(migrations, map[int]string{1: "one"})
	if err != nil || len(plan) != 1 || plan[0].Version != 2 {
		t.Fatalf("plan = %#v, err = %v", plan, err)
	}
	plan, err = Plan(migrations, map[int]string{1: "one", 2: "two"})
	if err != nil || len(plan) != 0 {
		t.Fatalf("idempotent plan = %#v, err = %v", plan, err)
	}
	if _, err := Plan(migrations, map[int]string{1: "changed"}); err == nil {
		t.Fatal("checksum change was accepted")
	}
	if _, err := Plan(migrations, map[int]string{3: "unknown"}); err == nil {
		t.Fatal("unknown applied version was accepted")
	}
}

func TestEmbeddedMigrationsAreNumbered(t *testing.T) {
	migrations, err := Embedded()
	if err != nil || len(migrations) != 1 || migrations[0].Version != 1 || migrations[0].Checksum == "" {
		t.Fatalf("embedded = %#v, err = %v", migrations, err)
	}
}
