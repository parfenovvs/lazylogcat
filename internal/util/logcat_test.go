package util

import "testing"

func TestParseProcessList(t *testing.T) {
	t.Run("NormalOutput", func(t *testing.T) {
		output := "PID NAME\n" +
			"  1234 com.example.app\n" +
			"  5678 com.other.app\n" +
			"   910 system_server\n"

		got := ParseProcessList(output)
		if len(got) != 3 {
			t.Fatalf("expected 3 processes, got %d", len(got))
		}

		expected := []Process{
			{PID: "1234", Name: "com.example.app"},
			{PID: "5678", Name: "com.other.app"},
			{PID: "910", Name: "system_server"},
		}
		for i, want := range expected {
			if got[i].PID != want.PID || got[i].Name != want.Name {
				t.Errorf("process[%d] = %+v, want %+v", i, got[i], want)
			}
		}
	})

	t.Run("EmptyOutput", func(t *testing.T) {
		got := ParseProcessList("")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("HeaderOnly", func(t *testing.T) {
		got := ParseProcessList("PID NAME\n")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("MalformedLines", func(t *testing.T) {
		output := "PID NAME\n" +
			"1234 com.example.app\n" +
			"incomplete\n" +
			"5678 com.other.app\n"

		got := ParseProcessList(output)
		if len(got) != 2 {
			t.Fatalf("expected 2 processes, got %d", len(got))
		}
		if got[0].PID != "1234" || got[0].Name != "com.example.app" {
			t.Errorf("process[0] = %+v, want {1234 com.example.app}", got[0])
		}
		if got[1].PID != "5678" || got[1].Name != "com.other.app" {
			t.Errorf("process[1] = %+v, want {5678 com.other.app}", got[1])
		}
	})

	t.Run("CarriageReturn", func(t *testing.T) {
		output := "PID NAME\r\n" +
			"1234 com.example.app\r\n"

		got := ParseProcessList(output)
		if len(got) != 1 {
			t.Fatalf("expected 1 process, got %d", len(got))
		}
		if got[0].PID != "1234" || got[0].Name != "com.example.app" {
			t.Errorf("process[0] = %+v, want {1234 com.example.app}", got[0])
		}
	})
}

func TestResolvePIDs(t *testing.T) {
	processes := []Process{
		{PID: "100", Name: "com.example.myapp"},
		{PID: "200", Name: "com.example.other"},
		{PID: "300", Name: "com.different.app"},
		{PID: "400", Name: "system_server"},
	}

	t.Run("EmptyFilter", func(t *testing.T) {
		got := ResolvePIDs(processes, "")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("SingleMatch", func(t *testing.T) {
		got := ResolvePIDs(processes, "myapp")
		if len(got) != 1 {
			t.Fatalf("expected 1 PID, got %d", len(got))
		}
		if _, ok := got["100"]; !ok {
			t.Errorf("expected PID 100 in set, got %v", got)
		}
	})

	t.Run("MultipleMatches", func(t *testing.T) {
		got := ResolvePIDs(processes, "example")
		if len(got) != 2 {
			t.Fatalf("expected 2 PIDs, got %d", len(got))
		}
		if _, ok := got["100"]; !ok {
			t.Errorf("expected PID 100 in set")
		}
		if _, ok := got["200"]; !ok {
			t.Errorf("expected PID 200 in set")
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		got := ResolvePIDs(processes, "MYAPP")
		if len(got) != 1 {
			t.Fatalf("expected 1 PID, got %d", len(got))
		}
		if _, ok := got["100"]; !ok {
			t.Errorf("expected PID 100 in set, got %v", got)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		got := ResolvePIDs(processes, "nonexistent")
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("MatchAll", func(t *testing.T) {
		got := ResolvePIDs(processes, "com.")
		if len(got) != 3 {
			t.Fatalf("expected 3 PIDs, got %d", len(got))
		}
	})

	t.Run("EmptyProcessList", func(t *testing.T) {
		got := ResolvePIDs(nil, "test")
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})
}
